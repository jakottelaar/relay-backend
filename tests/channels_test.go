package tests

import (
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/users"
	"github.com/stretchr/testify/assert"
)

func TestCreateChannel(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	user1 := createTestUser(t, app, users.RegisterRequest{
		Username: "test-username",
		Email:    "test-user@mail.com",
		Password: "test-password",
	})

	tests := []struct {
		name       string
		payload    map[string]any
		token      string
		wantStatus int
	}{
		{
			name: "valid channel",
			payload: map[string]any{
				"name":         "test-channel",
				"channel_type": "dm",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			payload: map[string]any{
				"channel_type": "dm",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing channel type",
			payload: map[string]any{
				"name": "test-channel",
			},
			token:      user1.AccessToken,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid channel type",
			payload: map[string]any{
				"name":         "test-channel",
				"channel_type": "invalid",
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

			w := performRequest(t, app, http.MethodPost, "/api/v1/channels", tt.payload, headers)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}

}
