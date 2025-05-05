package messages

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type MessagesRepo interface {
	SaveMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error)
}

type messagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) MessagesRepo {
	return &messagesRepo{db: db}
}

func (r *messagesRepo) SaveMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error) {
	query := `
		INSERT INTO messages (sender_id, channel_id, content) 
		VALUES ($1, $2, $3) 
		RETURNING id, sender_id, channel_id, content, created_at, updated_at
	`

	var message Message
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, senderID, channelID, content).Scan(
		&message.ID, &message.SenderID, &message.ChannelID, &message.Content, &message.CreatedAt, &message.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
