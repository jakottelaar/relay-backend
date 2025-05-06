package messages

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal/common"
)

type MessagesRepo interface {
	SaveMessage(ctx context.Context, senderID, channelID uuid.UUID, content string) (*Message, error)
	FindMessages(ctx context.Context, userID, channelID uuid.UUID, filters common.Filters) ([]*Message, common.Metadata, error)
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

func (r *messagesRepo) FindMessages(ctx context.Context, userID, channelID uuid.UUID, filters common.Filters) ([]*Message, common.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, sender_id, channel_id, content, created_at, updated_at, deleted_at, is_edited 
		FROM messages 
		WHERE channel_id = $1 AND sender_id = $2 
		ORDER BY %s %s
		LIMIT $3 OFFSET $4
	`, filters.SortColumn(), filters.SortDirection())

	rows, err := r.db.QueryContext(ctx, query, channelID, userID, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, common.Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var messages []*Message

	for rows.Next() {
		var message Message
		if err := rows.Scan(&totalRecords, &message.ID, &message.SenderID, &message.ChannelID, &message.Content,
			&message.CreatedAt, &message.UpdatedAt, &message.DeletedAt, &message.IsEdited); err != nil {
			return nil, common.Metadata{}, err
		}
		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, common.Metadata{}, err
	}

	metadata := common.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return messages, metadata, nil
}
