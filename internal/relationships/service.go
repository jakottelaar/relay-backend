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
	GetAllRelationships(ctx context.Context, current_user_id uuid.UUID) ([]*Relationship, error)
	AcceptFriendRequest(ctx context.Context, current_user_id uuid.UUID, other_user_id uuid.UUID) (*Relationship, error)
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
		return nil, internal.NewNotFoundError("User not found")
	}

	if targetUser.ID == current_user_id {
		return nil, internal.NewBadRequestError("Cannot send friend request to self")
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

func (s *relationshipsService) GetAllRelationships(ctx context.Context, current_user_id uuid.UUID) ([]*Relationship, error) {
	relationships, err := s.relationshipsRepo.FindAllRelationshipsByUserID(ctx, current_user_id)
	if err != nil {
		return nil, fmt.Errorf("could not get relationships: %w", err)
	}

	return relationships, nil
}

func (s *relationshipsService) AcceptFriendRequest(ctx context.Context, current_user_id uuid.UUID, other_user_id uuid.UUID) (*Relationship, error) {
	// Fetch target user to ensure they exist
	targetUser, err := s.usersRepo.FindUserByID(ctx, other_user_id.String())
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, internal.NewNotFoundError("User not found")
	}

	// Prevent accepting a friend request from yourself
	if targetUser.ID == current_user_id {
		return nil, internal.NewBadRequestError("Cannot accept own friend request")
	}

	// Fetch the relationship record
	relationship, err := s.relationshipsRepo.FindRelationshipByUserIDAndOtherUserID(ctx, current_user_id, other_user_id)
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
	updatedRelationship, err := s.relationshipsRepo.UpdateRelationshipStatus(ctx, current_user_id, other_user_id, RelationshipStatusFriend)
	if err != nil {
		return nil, fmt.Errorf("could not accept friend request: %w", err)
	}

	_, err = s.relationshipsRepo.UpdateRelationshipStatus(ctx, other_user_id, current_user_id, RelationshipStatusFriend)
	if err != nil {
		return nil, fmt.Errorf("could not update other user's relationship: %w", err)
	}

	return updatedRelationship, nil
}
