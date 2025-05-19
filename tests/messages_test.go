package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/channels"
	"github.com/jakottelaar/relay-backend/internal/messages"
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
		{
			name:       "invalid channel ID",
			userID:     user1.ID.String(),
			channelID:  "invalid-channel-id",
			wantStatus: http.StatusBadRequest,
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

func TestUpdateMessage(t *testing.T) {
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
	firstMessage := performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "Hello, world!"}, user1.ID.String())
	secondMessage := performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "How are you?"}, user2.ID.String())

	type MessageResponse struct {
		Message struct {
			messages.CreateMessageResponse
		} `json:"message"`
	}

	var firstMessageResp MessageResponse
	var secondMessageResp MessageResponse

	err = json.Unmarshal(firstMessage.Body.Bytes(), &firstMessageResp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	err = json.Unmarshal(secondMessage.Body.Bytes(), &secondMessageResp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	firstMessageID := firstMessageResp.Message.ID.String()

	tests := []struct {
		name       string
		userID     string
		messageID  string
		payload    map[string]any
		wantStatus int
	}{
		{
			name:       "valid request",
			userID:     user1.ID.String(),
			messageID:  firstMessageID,
			payload:    map[string]any{"content": "Updated message!"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid message ID",
			userID:     user1.ID.String(),
			messageID:  "invalid-message-id",
			payload:    map[string]any{"content": "Updated message!"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not the message sender",
			userID:     user2.ID.String(),
			messageID:  firstMessageID,
			payload:    map[string]any{"content": "Updated message!"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing content",
			userID:     user1.ID.String(),
			messageID:  firstMessageID,
			payload:    map[string]any{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty content",
			userID:     user1.ID.String(),
			messageID:  firstMessageID,
			payload:    map[string]any{"content": ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "message not found",
			userID:     user1.ID.String(),
			messageID:  uuid.New().String(),
			payload:    map[string]any{"content": "Updated message!"},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(t, testSetup.App, http.MethodPatch, "/api/v1/channels/"+channelID+"/messages/"+tt.messageID, tt.payload, tt.userID)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestDeleteMessage(t *testing.T) {
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
	firstMessage := performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "Hello, world!"}, user1.ID.String())
	secondMessage := performRequest(t, testSetup.App, http.MethodPost, "/api/v1/channels/"+channelID+"/messages", map[string]any{"content": "How are you?"}, user2.ID.String())

	type MessageResponse struct {
		Message struct {
			messages.CreateMessageResponse
		} `json:"message"`
	}

	var firstMessageResp MessageResponse
	var secondMessageResp MessageResponse

	err = json.Unmarshal(firstMessage.Body.Bytes(), &firstMessageResp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	err = json.Unmarshal(secondMessage.Body.Bytes(), &secondMessageResp)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	firstMessageID := firstMessageResp.Message.ID.String()
	secondMessageID := secondMessageResp.Message.ID.String()
	tests := []struct {
		name       string
		userID     string
		messageID  string
		wantStatus int
	}{
		{
			name:       "valid request",
			userID:     user1.ID.String(),
			messageID:  firstMessageID,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid message ID",
			userID:     user1.ID.String(),
			messageID:  "invalid-message-id",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not the message sender",
			userID:     user1.ID.String(),
			messageID:  secondMessageID,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "message not found",
			userID:     user1.ID.String(),
			messageID:  uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(t, testSetup.App, http.MethodDelete, "/api/v1/channels/"+channelID+"/messages/"+tt.messageID, nil, tt.userID)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
