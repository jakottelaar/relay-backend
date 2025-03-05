package tests

import (
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/users"
	"github.com/stretchr/testify/assert"
)

func TestCreateFriendRequest(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	createTestUser(t, app, users.RegisterRequest{
		Username: "test-username",
		Email:    "test-user@mail.com",
		Password: "test-password",
	})

	createTestUser(t, app, users.RegisterRequest{
		Username: "test-username2",
		Email:    "test-user2@mail.com",
		Password: "test-password",
	})

	tokenUser1 := loginUser(t, app, users.LoginRequest{
		Email:    "test-user@mail.com",
		Password: "test-password",
	})

	tokenUser2 := loginUser(t, app, users.LoginRequest{
		Email:    "test-user2@mail.com",
		Password: "test-password",
	})

	tests := []struct {
		name       string
		payload    map[string]interface{}
		token      string
		wantStatus int
	}{
		{
			name: "valid friend request",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			token:      tokenUser1,
			wantStatus: http.StatusCreated,
		},
		{
			name: "request to self",
			payload: map[string]interface{}{
				"username": "test-username",
			},
			token:      tokenUser1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			payload: map[string]interface{}{
				"username": "test-username3",
			},
			token:      tokenUser1,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "already outgoing friend request",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			token:      tokenUser1,
			wantStatus: http.StatusConflict,
		},
		{
			name: "user 2 sends friend request to user 1 while user 1 has outgoing request",
			payload: map[string]interface{}{
				"username": "test-username",
			},
			token:      tokenUser2,
			wantStatus: http.StatusCreated,
		},
		{
			name: "already friends",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			token:      tokenUser1,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": "Bearer " + tt.token,
			}

			w := performRequest(t, app, http.MethodPost, "/api/v1/relationships", tt.payload, headers)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
