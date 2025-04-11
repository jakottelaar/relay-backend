package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal/infra"
	"github.com/jakottelaar/relay-backend/internal/supabase"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestUser struct {
	ID       uuid.UUID
	Username string
	Email    string
}

type TestSetup struct {
	App                *infra.App
	MockAuthService    *MockAuthService
	MockSupabaseClient *MockSupabaseClient
	MockUsers          map[string]*TestUser
	Cleanup            func()
}

func SetupTestApp(t *testing.T) *TestSetup {
	ctx := context.Background()

	// Start test containers
	postgres, postgresCleanup := startPostgresContainer(t, ctx)
	postgresHost, err := postgres.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	postgresPort, err := postgres.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}
	postgresDSN := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", postgresHost, postgresPort.Port())

	// Run migrations
	if err := runMigrations(postgresDSN); err != nil {
		t.Fatalf("Failed to run migrations for PostgreSQL: %v", err)
	}

	// Create config
	cfg := &config.Config{
		Environment:       "test",
		Port:              8080,
		DSN:               postgresDSN,
		SupabaseJwtSecret: "test_secret",
		SupabaseUrl:       "http://mock-supabase",
		SupabaseApiKey:    "test-api-key",
	}

	// Create mock dependencies
	mockAuthService := new(MockAuthService)
	mockAuthMiddlewareProvider := new(MockAuthMiddlewareProvider)
	mockSupabaseClient := new(MockSupabaseClient)

	// Setup the mock auth middleware
	mockAuthMiddlewareProvider.On("AuthMiddleware").Return(gin.HandlerFunc(func(c *gin.Context) {
		userId := c.Request.Header.Get("X-Mock-User-ID")
		if userId == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("user_id", userId)
		c.Next()
	}))

	// Create app with mocked dependencies
	app, err := infra.NewApp(ctx, cfg, &infra.AppDependencies{
		AuthService:            mockAuthService,
		AuthMiddlewareProvider: mockAuthMiddlewareProvider,
		SupabaseClient:         mockSupabaseClient,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Create cleanup function
	cleanup := func() {
		app.Close()
		postgresCleanup()
	}

	return &TestSetup{
		App:                app,
		MockAuthService:    mockAuthService,
		MockSupabaseClient: mockSupabaseClient,
		MockUsers:          make(map[string]*TestUser),
		Cleanup:            cleanup,
	}
}

func startPostgresContainer(t *testing.T, ctx context.Context) (testcontainers.Container, func()) {
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

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		if err := postgres.Terminate(ctx); err != nil {
			t.Logf("Error terminating Postgres container: %v", err)
		}
	}

	return postgres, cleanup
}

func (ts *TestSetup) CreateMockUser(t *testing.T, username, email string) *TestUser {
	id := uuid.New()

	profile := &supabase.Profile{
		ID:        id,
		Username:  username,
		Email:     email,
		AvatarUrl: "https://example.com/avatar.png",
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Setup expectations for the mock Supabase client
	ts.MockSupabaseClient.On("GetUserByUsername", mock.Anything, username).Return(profile, nil)
	ts.MockSupabaseClient.On("GetUserByUsername", mock.Anything, "non-existent-user").
		Return(nil, errors.New("user not found"))

	ts.MockSupabaseClient.On("GetUserByID", mock.Anything, id).Return(profile, nil)
	ts.MockSupabaseClient.On("GetUserByID", mock.Anything, uuid.MustParse("00000000-0000-0000-0000-000000000000")).
		Return(nil, errors.New("user not found"))

	user := &TestUser{
		ID:       id,
		Username: username,
		Email:    email,
	}

	ts.MockUsers[username] = user
	return user
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

func performRequest(t *testing.T, app *infra.App, method, path string, body map[string]interface{}, userId string) *httptest.ResponseRecorder {
	jsonBody := ""
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(t, err)
		jsonBody = string(jsonBytes)
	}

	req, err := http.NewRequest(method, path, strings.NewReader(jsonBody))
	require.NoError(t, err)

	if userId != "" {
		req.Header.Set("X-Mock-User-ID", userId)
	}

	if jsonBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	app.HttpServer.Handler.ServeHTTP(w, req)
	return w
}
