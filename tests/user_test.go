package tests

import (
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/users"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	tests := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
	}{
		{
			name: "valid registration",
			payload: map[string]interface{}{
				"username": "test-username",
				"email":    "test-email@mail.com",
				"password": "test-password",
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := performRequest(t, app, http.MethodPost, "/api/v1/auth/register", tt.payload, nil)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Register user
	createTestMerchant(t, app, users.RegisterRequest{
		Username: "test-username",
		Email:    "test-user@mail.com",
		Password: "test-password",
	})

	tests := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
	}{
		{
			name: "valid login",
			payload: map[string]interface{}{
				"email":    "test-user@mail.com",
				"password": "test-password",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := performRequest(t, app, http.MethodPost, "/api/v1/auth/login", tt.payload, nil)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
