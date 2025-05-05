package messages

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
)

type MessagesHandler struct {
	service MessagesService
}

func NewMessagesHandler(service MessagesService) *MessagesHandler {
	return &MessagesHandler{service: service}
}

func (h *MessagesHandler) CreateMessage(c *gin.Context) {
	currentUserID, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		log.Printf("messages: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	channelID, err := uuid.Parse(c.Param("channel_id"))
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid channel id"))
		return
	}

	var request CreateMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("messages: failed to bind request: %v", err)
		_ = c.Error(internal.NewBadRequestError("Invalid request body"))
		return
	}

	message, err := h.service.CreateMessage(c.Request.Context(), userID, channelID, request.Content)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": &CreateMessageResponse{
		ID:        message.ID,
		SenderID:  message.SenderID,
		ChannelID: message.ChannelID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}})
}
