package channels

import (
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelTypeDM    ChannelType = "dm"
	ChannelTypeGroup ChannelType = "group"
)

type Channel struct {
	ID          string
	OwnerID     uuid.UUID
	Name        string
	ChannelType ChannelType
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChannelMember struct {
	ID        uuid.UUID
	ChannelID uuid.UUID
	UserID    uuid.UUID
	JoinedAt  time.Time
}

type GetChannelResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	OwnerID     uuid.UUID   `json:"owner_id"`
	ChannelType ChannelType `json:"channel_type"`
	CreatedAt   time.Time   `json:"created_at"`
}

type CreateGroupChannelRequest struct {
	Name    string   `json:"name" binding:"required"`
	UserIDs []string `json:"user_ids" binding:"required"`
}
