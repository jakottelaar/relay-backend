package tests

import (
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/infra"
	"github.com/jakottelaar/relay-backend/internal/users"
	"github.com/stretchr/testify/assert"
)

func sendFriendRequest(t *testing.T, app *infra.App, token, username string, wantStatus int) {
	w := performRequest(t, app, http.MethodPost, "/api/v1/relationships/friend-requests", map[string]interface{}{
		"username": username,
	}, map[string]string{
		"Authorization": "Bearer " + token,
	})
	assert.Equal(t, wantStatus, w.Code)
}

func acceptFriendRequest(t *testing.T, app *infra.App, token, otherUserID string, wantStatus int) {
	w := performRequest(t, app, http.MethodPatch, "/api/v1/relationships/users/"+otherUserID+"/friend-requests", nil, map[string]string{
		"Authorization": "Bearer " + token,
	})
	assert.Equal(t, wantStatus, w.Code)
}

func TestCreateFriendRequest(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	user1 := createTestUser(t, app, users.RegisterRequest{
		Username: "test-username",
		Email:    "test-user@mail.com",
		Password: "test-password",
	})

	user2 := createTestUser(t, app, users.RegisterRequest{
		Username: "test-username2",
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
			token:      user1.AccessToken,
			wantStatus: http.StatusCreated,
		},
		{
			name: "request to self",
			payload: map[string]interface{}{
				"username": "test-username",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			payload: map[string]interface{}{
				"username": "test-username3",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "already outgoing friend request",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusConflict,
		},
		{
			name: "user 2 sends friend request to user 1 while user 1 has outgoing request",
			payload: map[string]interface{}{
				"username": "test-username",
			},
			token:      user2.AccessToken,
			wantStatus: http.StatusCreated,
		},
		{
			name: "already friends",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "username not provided",
			payload:    map[string]interface{}{},
			token:      user1.AccessToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "too short username",
			payload: map[string]interface{}{
				"username": "a",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": "Bearer " + tt.token,
			}

			w := performRequest(t, app, http.MethodPost, "/api/v1/relationships/friend-requests", tt.payload, headers)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create users
	user1 := createTestUser(t, app, users.RegisterRequest{
		Username: "user1",
		Email:    "user1@mail.com",
		Password: "password",
	})

	user2 := createTestUser(t, app, users.RegisterRequest{
		Username: "user2",
		Email:    "user2@mail.com",
		Password: "password",
	})

	// User 1 sends friend request to User 2
	sendFriendRequest(t, app, user1.AccessToken, "user2", http.StatusCreated)

	tests := []struct {
		name        string
		token       string
		OtherUserID string
		wantStatus  int
	}{
		{
			name:        "valid accept friend request",
			token:       user2.AccessToken,
			OtherUserID: user1.ID.String(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "accept friend request that does not exist",
			token:       user2.AccessToken,
			OtherUserID: "00000000-0000-0000-0000-000000000000",
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "error: accept own friend request",
			token:       user2.AccessToken,
			OtherUserID: user2.ID.String(),
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": "Bearer " + tt.token,
			}

			w := performRequest(t, app, http.MethodPatch, "/api/v1/relationships/users/"+tt.OtherUserID+"/friend-requests", nil, headers)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestCancelOrDeclineFriendRequest(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create users
	user1 := createTestUser(t, app, users.RegisterRequest{
		Username: "user1",
		Email:    "user1@mail.com",
		Password: "password",
	})

	user2 := createTestUser(t, app, users.RegisterRequest{
		Username: "user2",
		Email:    "user2@mail.com",
		Password: "password",
	})

	// User 1 sends friend request to User 2
	sendFriendRequest(t, app, user1.AccessToken, "user2", http.StatusCreated)

	tests := []struct {
		name        string
		token       string
		OtherUserID string
		wantStatus  int
	}{
		{
			name:        "valid cancel/decline friend request",
			token:       user2.AccessToken,
			OtherUserID: user1.ID.String(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "error: cancel/decline friend request that does not exist",
			token:       user2.AccessToken,
			OtherUserID: "00000000-0000-0000-0000-000000000000",
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "error: cancel/decline own friend request",
			token:       user2.AccessToken,
			OtherUserID: user2.ID.String(),
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": "Bearer " + tt.token,
			}

			w := performRequest(t, app, http.MethodDelete, "/api/v1/relationships/users/"+tt.OtherUserID+"/friend-requests", nil, headers)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}

		})

	}
}

func TestRemoveFriend(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create users
	user1 := createTestUser(t, app, users.RegisterRequest{
		Username: "user1",
		Email:    "user1@mail.com",
		Password: "password",
	})

	user2 := createTestUser(t, app, users.RegisterRequest{
		Username: "user2",
		Email:    "user2@mail.com",
		Password: "password",
	})

	// User 1 sends friend request to User 2
	sendFriendRequest(t, app, user1.AccessToken, "user2", http.StatusCreated)

	// User 2 accepts friend request
	acceptFriendRequest(t, app, user2.AccessToken, user1.ID.String(), http.StatusOK)

	tests := []struct {
		name        string
		token       string
		OtherUserID string
		wantStatus  int
	}{
		{
			name:        "valid remove friend",
			token:       user1.AccessToken,
			OtherUserID: user2.ID.String(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "error: remove friend that does not exist",
			token:       user1.AccessToken,
			OtherUserID: "00000000-0000-0000-0000-000000000000",
			wantStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": "Bearer " + tt.token,
			}

			w := performRequest(t, app, http.MethodDelete, "/api/v1/relationships/users/"+tt.OtherUserID+"/friends", nil, headers)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
