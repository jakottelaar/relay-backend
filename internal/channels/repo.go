package channels

import (
	"context"
	"database/sql"
	"time"
)

type ChannelsRepo interface {
	SaveChannel(ctx context.Context, channel *Channel) (*Channel, error)
}

type channelsRepo struct {
	db *sql.DB
}

func NewChannelsRepo(db *sql.DB) ChannelsRepo {
	return &channelsRepo{db: db}
}

func (r *channelsRepo) SaveChannel(ctx context.Context, channel *Channel) (*Channel, error) {
	query := `
        INSERT INTO channels (name, owner_id, type)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx, query, channel.Name, channel.OwnerID, channel.ChannelType)
	err := row.Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return channel, nil
}
