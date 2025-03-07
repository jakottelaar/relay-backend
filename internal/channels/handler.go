package channels

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
)

type ChannelsHandler struct {
	service ChannelsService
}

func NewChannelsHandler(service ChannelsService) *ChannelsHandler {
	return &ChannelsHandler{service: service}
}

func (h *ChannelsHandler) CreateChannel(c *gin.Context) {
	currentUserId, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userId, err := uuid.Parse(currentUserId.(string))
	if err != nil {
		log.Printf("channels: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid request body"))
		return
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		_ = c.Error(internal.NewUnprocessableEntityError("Invalid input: " + err.Error()))
		return
	}

	channel, err := h.service.CreateChannel(c.Request.Context(), req.Name, userId, req.ChannelType)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"channel": &CreateChannelResponse{
			ID:          channel.ID,
			Name:        channel.Name,
			OwnerID:     channel.OwnerID,
			ChannelType: channel.ChannelType,
			CreatedAt:   channel.CreatedAt,
		},
	})
}
