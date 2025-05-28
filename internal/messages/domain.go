package messages

import (
	"time"

	"github.com/google/uuid"
)

const (
	SubjectMessageCreate = "messages.create"
	SubjectMessageUpdate = "messages.update"
	SubjectMessageDelete = "messages.delete"
)

type Message struct {
	ID        uuid.UUID
	SenderID  uuid.UUID
	ChannelID uuid.UUID
	Content   string
	IsEdited  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required" validate:"min=1,max=4096"`
}

type CreateMessageResponse struct {
	ID        uuid.UUID `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateMessageEvent struct {
	SenderID  uuid.UUID `json:"sender_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
}

type GetMessageResponse struct {
	ID        uuid.UUID  `json:"id"`
	SenderID  uuid.UUID  `json:"sender_id"`
	ChannelID uuid.UUID  `json:"channel_id"`
	Content   string     `json:"content"`
	IsEdited  bool       `json:"is_edited"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required" validate:"min=1,max=4096"`
}

type UpdateMessageResponse struct {
	ID        uuid.UUID `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
	IsEdited  bool      `json:"is_edited"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateMessageEvent struct {
	ID        uuid.UUID `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Content   string    `json:"content"`
}

type DeleteMessageEvent struct {
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
}
