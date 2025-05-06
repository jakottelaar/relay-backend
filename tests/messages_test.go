package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/channels"
	"github.com/stretchr/testify/assert"
)

func TestCreateMessage(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	// Create users
	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	// Create channel
	createdChannel := performRequest(t, testSetup.App, http.MethodGet, "/api/v1/users/"+user2.ID.String()+"/dm", nil, user1.ID.String())
	if createdChannel.Code != http.StatusOK {
		t.Errorf("Expected status %d but got %d: %s", http.StatusOK, createdChannel.Code, createdChannel.Body.String())
		return
	}

	channelResp := createdChannel.Body.String()
	var channelID string

	type ChannelResponse struct {
		Channel struct {
			channels.GetChannelResponse
		} `json:"channel"`
	}

	var resp ChannelResponse
	err := json.Unmarshal([]byte(channelResp), &resp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	channelID = resp.Channel.ID

	tests := []struct {
		name       string
		userID     string
		channelID  string
		payload    map[string]any
		wantStatus int
	}{
		{
			name:       "valid request",
			userID:     user1.ID.String(),
			channelID:  channelID,
			payload:    map[string]any{"content": "Hello, world!"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid channel ID",
			userID:     user1.ID.String(),
			channelID:  "invalid-channel-id",
			payload:    map[string]any{"content": "Hello, world!"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing content",
			userID:     user1.ID.String(),
			channelID:  channelID,
			payload:    map[string]any{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+tt.channelID+"/messages", tt.payload, tt.userID)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	// Create users
	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	// Create channel
	createdChannel := performRequest(t, testSetup.App, http.MethodGet, "/api/v1/users/"+user2.ID.String()+"/dm", nil, user1.ID.String())
	if createdChannel.Code != http.StatusOK {
		t.Errorf("Expected status %d but got %d: %s", http.StatusOK, createdChannel.Code, createdChannel.Body.String())
		return
	}

	channelResp := createdChannel.Body.String()
	var channelID string

	type ChannelResponse struct {
		Channel struct {
			channels.GetChannelResponse
		} `json:"channel"`
	}

	var resp ChannelResponse
	err := json.Unmarshal([]byte(channelResp), &resp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	channelID = resp.Channel.ID

	// Create messages
	performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "Hello, world!"}, user1.ID.String())
	performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "How are you?"}, user2.ID.String())
	performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "Goodbye!"}, user1.ID.String())
	performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "See you later!"}, user2.ID.String())

	tests := []struct {
		name       string
		userID     string
		channelID  string
		wantStatus int
	}{
		{
			name:       "valid request",
			userID:     user1.ID.String(),
			channelID:  channelID,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(t, testSetup.App, http.MethodGet, "/api/v1/channels/"+tt.channelID+"/messages", nil, tt.userID)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
