package relationships

import (
	"time"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/supabase"
)

const (
	SubjectRelationshipCreate   = "relationships.create"
	SubjectRelationshipAccepted = "relationships.accepted"
)

type RelationshipStatus string

const (
	RelationshipStatusNone         RelationshipStatus = "none"
	RelationshipStatusUser         RelationshipStatus = "user"
	RelationshipStatusFriend       RelationshipStatus = "friend"
	RelationshipStatusIncoming     RelationshipStatus = "incoming"
	RelationshipStatusOutgoing     RelationshipStatus = "outgoing"
	RelationshipStatusBlocked      RelationshipStatus = "blocked"
	RelationshipStatusBlockedOther RelationshipStatus = "blocked_other"
)

type Relationship struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	OtherUserID        uuid.UUID
	RelationshipStatus RelationshipStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateRelationshipRequest struct {
	Username string `json:"username" binding:"required" validate:"min=3,max=64"`
}

type CreateRelationshipResponse struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	OtherUserID        uuid.UUID `json:"other_user_id"`
	RelationshipStatus string    `json:"relationship_status"`
	CreatedAt          time.Time `json:"created_at"`
}

type CreateRelationshipEvent struct {
	ID          uuid.UUID        `json:"id"`
	OtherUserID uuid.UUID        `json:"other_user_id"`
	Sender      supabase.Profile `json:"sender"`
}

type GetRelationshipResponse struct {
	ID                 uuid.UUID         `json:"id"`
	UserID             uuid.UUID         `json:"user_id"`
	OtherUserID        uuid.UUID         `json:"other_user_id"`
	RelationshipStatus string            `json:"relationship_status"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	OtherUser          *supabase.Profile `json:"other_user,omitempty"`
}

type AcceptRelationshipEvent struct {
	ID          uuid.UUID        `json:"id"`
	OtherUserID uuid.UUID        `json:"other_user_id"`
	Sender      supabase.Profile `json:"sender"`
}
