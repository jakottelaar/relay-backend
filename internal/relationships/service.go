package relationships

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/supabase"
	"github.com/jakottelaar/relay-backend/internal/websocket"
)

type RelationshipsService interface {
	CreateRelationship(ctx context.Context, username string, currentUserID uuid.UUID) (*Relationship, error)
	GetAllRelationships(ctx context.Context, currentUserID uuid.UUID) ([]*GetRelationshipResponse, error)
	AcceptFriendRequest(ctx context.Context, currentUserID uuid.UUID, targetUserID uuid.UUID) (*Relationship, error)
	CancelOrRejectFriendRequest(ctx context.Context, currentUserID uuid.UUID, targetUserID uuid.UUID) (string, error)
	RemoveFriend(ctx context.Context, currentUserID uuid.UUID, targetUserID uuid.UUID) error
}

type relationshipsService struct {
	relationshipsRepo RelationshipsRepo
	supabaseClient    supabase.SupabaseClient
	wsManager         *websocket.Manager
}

func NewRelationshipsService(relationshipsRepo RelationshipsRepo, supabaseClient supabase.SupabaseClient, wsManager *websocket.Manager) RelationshipsService {
	return &relationshipsService{
		relationshipsRepo: relationshipsRepo,
		supabaseClient:    supabaseClient,
		wsManager:         wsManager,
	}
}

func (s *relationshipsService) CreateRelationship(ctx context.Context, username string, currentUserID uuid.UUID) (*Relationship, error) {
	targetUser, err := s.supabaseClient.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, internal.NewNotFoundError("User not found")
	}

	if targetUser.ID == currentUserID {
		return nil, internal.NewBadRequestError("Cannot send friend request to self")
	}

	existingRelationship, err := s.relationshipsRepo.FindRelationshipByUserIDAndTargetUserID(ctx, currentUserID, targetUser.ID)
	if err == nil {
		// Explicitly check the relationship status for the current user
		var currentUserStatus, otherUserStatus RelationshipStatus
		if existingRelationship.UserID == currentUserID {
			currentUserStatus = existingRelationship.RelationshipStatus
			// Swap the statuses to get the correct perspective
			otherUserStatus = s.getOppositeStatus(currentUserStatus)
		} else {
			otherUserStatus = existingRelationship.RelationshipStatus
			currentUserStatus = s.getOppositeStatus(otherUserStatus)
		}

		switch currentUserStatus {
		case RelationshipStatusFriend:
			return nil, internal.NewDuplicateError("Already friends")
		case RelationshipStatusOutgoing:
			return nil, internal.NewDuplicateError("Friend request already sent")
		case RelationshipStatusIncoming:
			// Check if the other user has an incoming request
			if otherUserStatus == RelationshipStatusOutgoing {
				// Auto-accept friend request
				relationship, err := s.relationshipsRepo.UpdateRelationshipStatus(ctx, currentUserID, targetUser.ID, RelationshipStatusFriend)
				if err != nil {
					return nil, fmt.Errorf("could not accept friend request: %w", err)
				}
				return relationship, nil
			}
		}

		return nil, internal.NewInternalServerError("Unexpected relationship state")
	}

	savedRelationship, err := s.relationshipsRepo.SaveRelationship(ctx, currentUserID, targetUser.ID)
	if err != nil {
		return nil, fmt.Errorf("could not save relationship: %w", err)
	}

	// Notify the other user about the new friend request
	senderProfile, err := s.supabaseClient.GetUserByID(ctx, currentUserID)
	if err != nil {
		// Log error but don't fail the request
		log.Printf("Error fetching sender profile: %v", err)
	} else {
		// Send WebSocket notification
		notification := map[string]any{
			"type": "FRIEND_REQUEST_RECEIVED",
			"data": map[string]any{
				"relationship_id": savedRelationship.ID.String(),
				"sender": map[string]any{
					"id":         senderProfile.ID.String(),
					"username":   senderProfile.Username,
					"avatar_url": senderProfile.AvatarUrl,
				},
			},
		}

		notificationJSON, _ := json.Marshal(notification)
		s.wsManager.SendToUser(targetUser.ID, notificationJSON)
	}

	return savedRelationship, nil
}

func (s *relationshipsService) getOppositeStatus(status RelationshipStatus) RelationshipStatus {
	switch status {
	case RelationshipStatusOutgoing:
		return RelationshipStatusIncoming
	case RelationshipStatusIncoming:
		return RelationshipStatusOutgoing
	default:
		return status
	}
}

func (s *relationshipsService) GetAllRelationships(ctx context.Context, currentUserID uuid.UUID) ([]*GetRelationshipResponse, error) {
	relationships, err := s.relationshipsRepo.FindAllRelationshipsByUserID(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("could not get relationships: %w", err)
	}

	otherUserIDs := make([]uuid.UUID, len(relationships))
	for i, rel := range relationships {
		otherUserIDs[i] = rel.OtherUserID
	}

	profilesMap, err := s.supabaseClient.GetUsersByIDs(ctx, otherUserIDs)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user profiles: %w", err)
	}

	result := make([]*GetRelationshipResponse, len(relationships))
	for i, rel := range relationships {
		result[i] = &GetRelationshipResponse{
			ID:                 rel.ID,
			UserID:             rel.UserID,
			OtherUserID:        rel.OtherUserID,
			RelationshipStatus: string(rel.RelationshipStatus),
			CreatedAt:          rel.CreatedAt,
			UpdatedAt:          rel.UpdatedAt,
			OtherUser:          profilesMap[rel.OtherUserID],
		}
	}

	return result, nil
}

func (s *relationshipsService) AcceptFriendRequest(ctx context.Context, currentUserID uuid.UUID, targetUserID uuid.UUID) (*Relationship, error) {
	// Fetch target user to ensure they exist
	targetUser, err := s.supabaseClient.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, internal.NewNotFoundError("User not found")
	}

	// Prevent accepting a friend request from yourself
	if targetUser.ID == currentUserID {
		return nil, internal.NewBadRequestError("Cannot accept own friend request")
	}

	// Fetch the relationship record
	relationship, err := s.relationshipsRepo.FindRelationshipByUserIDAndTargetUserID(ctx, currentUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("could not find relationship: %w", err)
	}
	if relationship == nil {
		return nil, internal.NewBadRequestError("No friend request found")
	}

	// Check if they are already friends
	if relationship.RelationshipStatus == RelationshipStatusFriend {
		return nil, internal.NewBadRequestError("Already friends")
	}

	// Ensure the friend request exists in the correct state
	if relationship.RelationshipStatus == RelationshipStatusOutgoing {
		return nil, internal.NewBadRequestError("Cannot accept outgoing friend request")
	}

	// Ensure the friend request exists in the correct state
	if relationship.RelationshipStatus != RelationshipStatusIncoming {
		return nil, internal.NewBadRequestError("Unexpected relationship state")
	}

	// Update both records to "friend"
	updatedRelationship, err := s.relationshipsRepo.UpdateRelationshipStatus(ctx, currentUserID, targetUserID, RelationshipStatusFriend)
	if err != nil {
		return nil, fmt.Errorf("could not accept friend request: %w", err)
	}

	_, err = s.relationshipsRepo.UpdateRelationshipStatus(ctx, targetUserID, currentUserID, RelationshipStatusFriend)
	if err != nil {
		return nil, fmt.Errorf("could not update other user's relationship: %w", err)
	}

	return updatedRelationship, nil
}

func (s *relationshipsService) CancelOrRejectFriendRequest(ctx context.Context, currentUserID uuid.UUID, targetUserID uuid.UUID) (string, error) {
	// Fetch target user to ensure they exist
	targetUser, err := s.supabaseClient.GetUserByID(ctx, targetUserID)
	if err != nil {
		return "", internal.NewNotFoundError("User not found")
	}

	// Prevent cancelling a friend request to yourself
	if targetUser.ID == currentUserID {
		return "", internal.NewBadRequestError("Cannot cancel own friend request")
	}

	// Fetch the relationship record
	relationship, err := s.relationshipsRepo.FindRelationshipByUserIDAndTargetUserID(ctx, currentUserID, targetUserID)
	if err != nil {
		return "", fmt.Errorf("could not find relationship: %w", err)
	}
	if relationship == nil {
		return "", internal.NewBadRequestError("No friend request found")
	}

	// Check if they are already friends
	if relationship.RelationshipStatus == RelationshipStatusFriend {
		return "", internal.NewBadRequestError("Already friends")
	}

	// Ensure the friend request exists in the correct state
	if relationship.RelationshipStatus == RelationshipStatusIncoming {
		// Delete the incoming relationship
		err = s.relationshipsRepo.DeleteRelationship(ctx, currentUserID, targetUserID)
		if err != nil {
			return "", fmt.Errorf("could not delete incoming relationship: %w", err)
		}
		return "Friend request declined", nil
	}

	// Ensure the friend request exists in the correct state
	if relationship.RelationshipStatus != RelationshipStatusOutgoing {
		return "", internal.NewBadRequestError("Unexpected relationship state")
	}

	// Delete the outgoing relationship
	err = s.relationshipsRepo.DeleteRelationship(ctx, currentUserID, targetUserID)
	if err != nil {
		return "", fmt.Errorf("could not delete outgoing relationship: %w", err)
	}

	return "Friend request cancelled", nil
}

func (s *relationshipsService) RemoveFriend(ctx context.Context, currentUserID uuid.UUID, targetUserID uuid.UUID) error {
	_, err := s.supabaseClient.GetUserByID(ctx, targetUserID)
	if err != nil {
		return internal.NewNotFoundError("User not found")
	}

	relationship, err := s.relationshipsRepo.FindRelationshipByUserIDAndTargetUserID(ctx, currentUserID, targetUserID)
	if err != nil {
		return fmt.Errorf("could not find relationship: %w", err)
	}

	if relationship == nil {
		return internal.NewBadRequestError("No relationship found")
	}

	if relationship.RelationshipStatus != RelationshipStatusFriend {
		return internal.NewBadRequestError("Not friends")
	}

	err = s.relationshipsRepo.DeleteRelationship(ctx, currentUserID, targetUserID)
	if err != nil {
		return fmt.Errorf("could not delete relationship: %w", err)
	}

	err = s.relationshipsRepo.DeleteRelationship(ctx, targetUserID, currentUserID)
	if err != nil {
		return fmt.Errorf("could not delete other user's relationship: %w", err)
	}

	return nil
}
