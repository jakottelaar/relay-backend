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
		{
			name: "missing username",
			payload: map[string]interface{}{
				"email":    "test-email2@mail.com",
				"password": "test-password",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			payload: map[string]interface{}{
				"username": "test-username2",
				"password": "test-password",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: map[string]interface{}{
				"username": "test-username3",
				"email":    "test-email3@mail.com",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			payload: map[string]interface{}{
				"username": "test-username4",
				"email":    "test-email4",
				"password": "test-password",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "short password",
			payload: map[string]interface{}{
				"username": "test-username5",
				"email":    "test-email5@mail.com",
				"password": "st",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "long password",
			payload: map[string]interface{}{
				"username": "test-username6",
				"email":    "test-email6@mail.com",
				"password": "thispasswordiswaytoolongandshouldnotbeacceptedABCDEFGHIJKLMNOPQRSTUVWXZABCDEFGHIJKLMNOPQRSTUVWXZ",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "duplicate email",
			payload: map[string]interface{}{
				"username": "test-username7",
				"email":    "test-email@mail.com",
				"password": "test-password",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "duplicate username",
			payload: map[string]interface{}{
				"username": "test-username",
				"email":    "test-email2@mail.com",
				"password": "test-password",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid payload",
			payload: map[string]interface{}{
				"username": "test-username",
				"email":    "test-email@mail.com",
				"password": 123,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty payload",
			payload:    map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
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
	createTestUser(t, app, users.RegisterRequest{
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
		{
			name: "missing email",
			payload: map[string]interface{}{
				"password": "test-password",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: map[string]interface{}{
				"email": "test-user@mail.com",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			payload: map[string]interface{}{
				"email":    "test-user",
				"password": "test-password",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "user not found",
			payload: map[string]interface{}{
				"email":    "test-nonexistinguser@mail.com",
				"password": "test-password",
			},
			wantStatus: http.StatusNotFound,
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
