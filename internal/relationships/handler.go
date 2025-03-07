package relationships

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
)

type RelationshipsHandler struct {
	service RelationshipsService
}

func NewRelationshipsHandler(service RelationshipsService) *RelationshipsHandler {
	return &RelationshipsHandler{service: service}
}

func (h *RelationshipsHandler) CreateRelationship(c *gin.Context) {
	var req CreateRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid request body"))
		return
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		_ = c.Error(internal.NewUnprocessableEntityError("Invalid input: " + err.Error()))
		return
	}

	currentUserID, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		log.Printf("relationships: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	relationship, err := h.service.CreateRelationship(c.Request.Context(), req.Username, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"relationship": &CreateRelationshipResponse{
		ID:                 relationship.ID,
		UserID:             relationship.UserID,
		OtherUserID:        relationship.OtherUserID,
		RelationshipStatus: string(relationship.RelationshipStatus),
		CreatedAt:          relationship.CreatedAt,
	}})

}

func (h *RelationshipsHandler) GetAllRelationships(c *gin.Context) {
	currentUserID, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		log.Printf("relationships: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	relationships, err := h.service.GetAllRelationships(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var relationshipsResponse []*GetRelationshipResponse
	for _, relationship := range relationships {
		relationshipsResponse = append(relationshipsResponse, &GetRelationshipResponse{
			ID:                 relationship.ID,
			UserID:             relationship.UserID,
			OtherUserID:        relationship.OtherUserID,
			RelationshipStatus: string(relationship.RelationshipStatus),
			CreatedAt:          relationship.CreatedAt,
			UpdatedAt:          relationship.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"relationships": relationshipsResponse})
}

func (h *RelationshipsHandler) AcceptFriendRequest(c *gin.Context) {
	currentUserID, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		log.Printf("relationships: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("target_user_id"))
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid relationship id"))
		return
	}

	_, err = h.service.AcceptFriendRequest(c.Request.Context(), userID, targetUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend request accepted",
	})
}

func (h *RelationshipsHandler) CancelOrDeclineFriendRequest(c *gin.Context) {
	currentUserID, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		log.Printf("relationships: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("target_user_id"))
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid relationship id"))
		return
	}

	message, err := h.service.CancelOrDeclineFriendRequest(c.Request.Context(), userID, targetUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

func (h *RelationshipsHandler) RemoveFriend(c *gin.Context) {

	currentUserID, ok := c.Get("user_id")
	if !ok {
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		log.Printf("relationships: failed to parse user_id: %v", err)
		_ = c.Error(internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("target_user_id"))
	if err != nil {
		_ = c.Error(internal.NewBadRequestError("Invalid target user id"))
		return
	}

	err = h.service.RemoveFriend(c.Request.Context(), userID, targetUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend removed",
	})

}
