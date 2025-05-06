package messages

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/common"
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

	c.JSON(http.StatusCreated, gin.H{"message": &CreateMessageResponse{
		ID:        message.ID,
		SenderID:  message.SenderID,
		ChannelID: message.ChannelID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}})
}

func (h *MessagesHandler) GetMessages(c *gin.Context) {
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

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	sort := c.DefaultQuery("sort", "-created_at")
	sortSafeList := []string{"created_at", "-created_at"}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid page number"))
		return
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid page size"))
		return
	}

	filters := common.Filters{
		Page:         pageInt,
		PageSize:     pageSizeInt,
		Sort:         sort,
		SortSafelist: sortSafeList,
	}

	messages, metadata, err := h.service.GetMessages(c.Request.Context(), userID, channelID, filters)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if len(messages) == 0 {
		c.JSON(http.StatusOK, gin.H{"messages": []string{}})
		return
	}

	var responseMessages []GetMessageResponse
	for _, message := range messages {
		responseMessages = append(responseMessages, GetMessageResponse{
			ID:        message.ID,
			SenderID:  message.SenderID,
			ChannelID: message.ChannelID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
			UpdatedAt: message.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"messages": responseMessages, "metadata": metadata})
}
