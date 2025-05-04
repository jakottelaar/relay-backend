package websocket

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal/auth"
)

const (
	MessageTypeFriendRequest = "FRIEND_REQUEST"
)

type WebSocketHandler struct {
	wsManager *Manager
	config    *config.Config
}

func NewWebSocketHandler(wsManager *Manager, config *config.Config) *WebSocketHandler {
	return &WebSocketHandler{
		wsManager: wsManager,
		config:    config,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	authService := &auth.SupabaseAuthService{
		JwtSecret: h.config.SupabaseJwtSecret,
	}

	currentUserId, err := auth.AuthenticateWebSocketRequest(c, authService)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if errors.Is(err, auth.ErrTokenExpired) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		return
	}

	h.wsManager.GetMelody().HandleRequestWithKeys(c.Writer, c.Request, map[string]any{
		"user_id": currentUserId,
	})
}
