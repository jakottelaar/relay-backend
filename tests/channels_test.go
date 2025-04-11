package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDMChannel(t *testing.T) {
	testSetup := SetupTestApp(t)
	defer testSetup.Cleanup()

	// Create users
	user1 := testSetup.CreateMockUser(t, "test-username", "test-user@mail.com")
	user2 := testSetup.CreateMockUser(t, "test-username2", "test-user2@mail.com")

	tests := []struct {
		name         string
		userID       string
		targetuserID string
		wantStatus   int
	}{
		{
			name:         "valid request",
			userID:       user1.ID.String(),
			targetuserID: user2.ID.String(),
			wantStatus:   http.StatusOK,
		},
		{
			name:         "invalid request",
			userID:       user1.ID.String(),
			targetuserID: "invalid-user-id",
			wantStatus:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := performRequest(t, testSetup.App, http.MethodGet, "/api/v1/users/"+tt.targetuserID+"/dm", nil, tt.userID)
			assert.Equal(t, tt.wantStatus, w.Code)
			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d but got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}

}
