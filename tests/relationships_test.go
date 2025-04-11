package tests

import (
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/infra"
	"github.com/stretchr/testify/assert"
)

func sendFriendRequest(t *testing.T, app *infra.App, userID, username string, wantStatus int) {
	w := performRequest(t, app, http.MethodPost, "/api/v1/relationships/friend-requests", map[string]interface{}{
		"username": username,
	}, userID)
	assert.Equal(t, wantStatus, w.Code)
}

func acceptFriendRequest(t *testing.T, app *infra.App, userID, otherUserID string, wantStatus int) {
	w := performRequest(t, app, http.MethodPatch, "/api/v1/relationships/users/"+otherUserID+"/friend-requests", nil, userID)
	assert.Equal(t, wantStatus, w.Code)
}

func TestCreateFriendRequest(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	tests := []struct {
		name       string
		payload    map[string]interface{}
		userID     string
		wantStatus int
	}{
		{
			name: "valid friend request",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			userID:     user1.ID.String(),
			wantStatus: http.StatusCreated,
		},
		{
			name: "request to self",
			payload: map[string]interface{}{
				"username": "test-username",
			},
			userID:     user1.ID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			payload: map[string]interface{}{
				"username": "non-existent-user",
			},
			userID:     user1.ID.String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name: "already outgoing friend request",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			userID:     user1.ID.String(),
			wantStatus: http.StatusConflict,
		},
		{
			name: "user 2 sends friend request to user 1 while user 1 has outgoing request",
			payload: map[string]interface{}{
				"username": "test-username",
			},
			userID:     user2.ID.String(),
			wantStatus: http.StatusCreated,
		},
		{
			name: "already friends",
			payload: map[string]interface{}{
				"username": "test-username2",
			},
			userID:     user1.ID.String(),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "username not provided",
			payload:    map[string]interface{}{},
			userID:     user1.ID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "too short username",
			payload: map[string]interface{}{
				"username": "a",
			},
			userID:     user1.ID.String(),
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(
				t,
				testSetup.App,
				http.MethodPost,
				"/api/v1/relationships/friend-requests",
				tt.payload,
				tt.userID,
			)

			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	// Create users
	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	// User 1 sends friend request to User 2
	sendFriendRequest(t, testSetup.App, user1.ID.String(), user2.Username, http.StatusCreated)

	tests := []struct {
		name        string
		userID      string
		OtherUserID string
		wantStatus  int
	}{
		{
			name:        "valid accept friend request",
			userID:      user2.ID.String(),
			OtherUserID: user1.ID.String(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "accept friend request that does not exist",
			userID:      user2.ID.String(),
			OtherUserID: "00000000-0000-0000-0000-000000000000",
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "error: accept own friend request",
			userID:      user2.ID.String(),
			OtherUserID: user2.ID.String(),
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := performRequest(
				t,
				testSetup.App,
				http.MethodPatch,
				"/api/v1/relationships/users/"+tt.OtherUserID+"/friend-requests",
				nil,
				tt.userID,
			)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestCancelOrDeclineFriendRequest(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	// Create users
	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	// User 1 sends friend request to User 2
	sendFriendRequest(t, testSetup.App, user1.ID.String(), user2.Username, http.StatusCreated)

	tests := []struct {
		name        string
		userID      string
		OtherUserID string
		wantStatus  int
	}{
		{
			name:        "valid cancel/decline friend request",
			userID:      user2.ID.String(),
			OtherUserID: user1.ID.String(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "error: cancel/decline friend request that does not exist",
			userID:      user2.ID.String(),
			OtherUserID: "00000000-0000-0000-0000-000000000000",
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "error: cancel/decline own friend request",
			userID:      user2.ID.String(),
			OtherUserID: user2.ID.String(),
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(
				t,
				testSetup.App,
				http.MethodDelete,
				"/api/v1/relationships/users/"+tt.OtherUserID+"/friend-requests",
				nil,
				tt.userID,
			)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}

		})

	}
}

func TestRemoveFriend(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	// Create users
	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	// User 1 sends friend request to User 2
	sendFriendRequest(t, testSetup.App, user1.ID.String(), user2.Username, http.StatusCreated)

	// User 2 accepts friend request
	acceptFriendRequest(t, testSetup.App, user2.ID.String(), user1.ID.String(), http.StatusOK)

	tests := []struct {
		name        string
		userID      string
		OtherUserID string
		wantStatus  int
	}{
		{
			name:        "valid remove friend",
			userID:      user1.ID.String(),
			OtherUserID: user2.ID.String(),
			wantStatus:  http.StatusOK,
		},
		{
			name:        "error: remove friend that does not exist",
			userID:      user1.ID.String(),
			OtherUserID: "00000000-0000-0000-0000-000000000000",
			wantStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := performRequest(t, testSetup.App, http.MethodDelete, "/api/v1/relationships/users/"+tt.OtherUserID+"/friends", nil, tt.userID)
			assert.Equal(t, tt.wantStatus, w.Code)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
