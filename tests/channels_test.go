package tests

import (
	"net/http"
	"testing"

	"github.com/jakottelaar/relay-backend/internal/users"
	"github.com/stretchr/testify/assert"
)

func TestCreateDMChannel(t *testing.T) {
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
		name         string
		token        string
		targetUserID string
		wantStatus   int
	}{
		{
			name:         "valid request",
			token:        user1.AccessToken,
			targetUserID: user2.ID.String(),
			wantStatus:   http.StatusOK,
		},
		{
			name:         "invalid request",
			token:        user1.AccessToken,
			targetUserID: "invalid-user-id",
			wantStatus:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": "Bearer " + tt.token,
			}

			w := performRequest(t, app, http.MethodGet, "/api/v1/users/"+tt.targetUserID+"/dm", nil, headers)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}

}
