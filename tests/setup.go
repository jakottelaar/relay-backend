package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal/infra"
	"github.com/jakottelaar/relay-backend/internal/users"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestApp(t *testing.T) (*infra.App, func()) {
	ctx := context.Background()

	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
	}

	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	redis, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	postgresHost, err := postgres.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	postgresPort, err := postgres.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	postgresDSN := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", postgresHost, postgresPort.Port())

	if err := runMigrations(postgresDSN); err != nil {
		t.Fatalf("Failed to run migrations for PostgreSQL: %v", err)
	}

	cfg := &config.Config{
		Environment:         "test",
		Port:                8080,
		DSN:                 postgresDSN,
		JwtSecret:           "test_secret",
		JwtExpirationSecond: 3600,
	}

	app, err := infra.NewApp(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		app.Close()
		if err := postgres.Terminate(ctx); err != nil {
			t.Logf("Error terminating Postgres container: %v", err)
		}
		if err := redis.Terminate(ctx); err != nil {
			t.Logf("Error terminating Redis container: %v", err)
		}
	}

	return app, cleanup
}

func runMigrations(dsn string) error {
	// Create a new migration instance
	m, err := migrate.New(
		"file://../migrations",
		dsn,
	)

	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Apply migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

func createTestMerchant(t *testing.T, app *infra.App, req users.RegisterRequest) {

	// Register user
	w := performRequest(t, app, http.MethodPost, "/api/v1/auth/register", req, nil)
	assert.Equal(t, http.StatusCreated, w.Code)

	var signInResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &signInResponse)
	if err != nil {
		t.Fatalf("Error unmarshalling sign-in response: %v", err)
	}
}

func performRequest(t *testing.T, app *infra.App, method, path string, payload interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var jsonBody []byte
	if payload != nil {
		var err error
		jsonBody, err = json.Marshal(payload)
		if err != nil {
			t.Fatalf("Error marshalling payload: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	app.HttpServer.Handler.ServeHTTP(w, req)
	return w
}
