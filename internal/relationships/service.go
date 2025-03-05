package relationships

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/users"
)

type RelationshipsService interface {
	CreateRelationship(ctx context.Context, username string, current_user_id uuid.UUID) (*Relationship, error)
}

type relationshipsService struct {
	relationshipsRepo RelationshipsRepo
	usersRepo         users.UserRepo
}

func NewRelationshipsService(relationshipsRepo RelationshipsRepo, usersRepo users.UserRepo) RelationshipsService {
	return &relationshipsService{
		relationshipsRepo: relationshipsRepo,
		usersRepo:         usersRepo,
	}
}

func (s *relationshipsService) CreateRelationship(ctx context.Context, username string, current_user_id uuid.UUID) (*Relationship, error) {
	targetUser, err := s.usersRepo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, internal.ErrUserNotFound
	}

	existingRelationship, err := s.relationshipsRepo.FindRelationshipByUserIDAndOtherUserID(ctx, current_user_id, targetUser.ID)
	if err == nil {
		// Explicitly check the relationship status for the current user
		var currentUserStatus, otherUserStatus RelationshipStatus
		if existingRelationship.UserID == current_user_id {
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
				relationship, err := s.relationshipsRepo.UpdateRelationshipStatus(ctx, current_user_id, targetUser.ID, RelationshipStatusFriend)
				if err != nil {
					return nil, fmt.Errorf("could not accept friend request: %w", err)
				}
				return relationship, nil
			}
		}

		return nil, internal.NewInternalServerError("Unexpected relationship state")
	}

	savedRelationship, err := s.relationshipsRepo.SaveRelationship(ctx, current_user_id, targetUser.ID)
	if err != nil {
		return nil, fmt.Errorf("could not save relationship: %w", err)
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
