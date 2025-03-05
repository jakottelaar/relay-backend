package relationships

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

	log.Printf("CreateRelationshipRequest: %v", req.Username)

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserID, ok := c.Get("user_id")
	log.Printf("currentUserID: %v", currentUserID)
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
