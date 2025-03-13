package channels

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
)

type ChannelsHandler struct {
	service ChannelsService
}

func NewChannelsHandler(service ChannelsService) *ChannelsHandler {
	return &ChannelsHandler{service: service}
}

func (h *ChannelsHandler) GetDMChannel(c *gin.Context) {
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

	targetUserID, err := uuid.Parse(c.Param("target_user_id"))
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid target user id"))
		return
	}

	channel, err := h.service.GetDMChannel(c.Request.Context(), userId, targetUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel": &GetChannelResponse{
			ID:          channel.ID,
			Name:        channel.Name,
			OwnerID:     channel.OwnerID,
			ChannelType: channel.ChannelType,
			CreatedAt:   channel.CreatedAt,
		},
	})
}

func (h *ChannelsHandler) CreateGroupChannel(c *gin.Context) {
	currentUserId, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	ownerUserID, err := uuid.Parse(currentUserId.(string))
	if err != nil {
		log.Printf("channels: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	var req CreateGroupChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid request body"))
		return
	}

	userIds := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, id := range req.UserIDs {
		userId, err := uuid.Parse(id)
		if err != nil {
			_ = c.Error(internal.NewBadRequestError("Invalid user id"))
			return
		}

		userIds = append(userIds, userId)
	}

	channel, err := h.service.CreateGroupChannel(c.Request.Context(), ownerUserID, req.Name, userIds)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"channel": &GetChannelResponse{
			ID:          channel.ID,
			Name:        channel.Name,
			OwnerID:     channel.OwnerID,
			ChannelType: channel.ChannelType,
			CreatedAt:   channel.CreatedAt,
		},
	})
}
