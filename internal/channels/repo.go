package channels

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jakottelaar/relay-backend/internal"
)

type ChannelsRepo interface {
	SaveChannel(ctx context.Context, channel *Channel, tx *sql.Tx) (*Channel, error)
	FindDMChannelByUserIDs(ctx context.Context, userID, targetUserID uuid.UUID) (*Channel, error)
	SaveDMChannel(ctx context.Context, userId, targetUserID uuid.UUID) (*Channel, error)
	AddUserToChannel(ctx context.Context, channelID, userID uuid.UUID, tx *sql.Tx) (uuid.UUID, error)
	SaveGroupChannel(ctx context.Context, ownerUserID uuid.UUID, name string, channelMemberIDs []uuid.UUID) (*Channel, []uuid.UUID, error)
	FindAllChannelsByUserID(ctx context.Context, userID uuid.UUID) ([]*ChannelWithMembers, error)
	FindDMChannelByID(ctx context.Context, channelID uuid.UUID) (*Channel, error)
}

type channelsRepo struct {
	db *sql.DB
}

func NewChannelsRepo(db *sql.DB) ChannelsRepo {
	return &channelsRepo{db: db}
}

func (r *channelsRepo) SaveChannel(ctx context.Context, channel *Channel, tx *sql.Tx) (*Channel, error) {
	query := `
        INSERT INTO channels (name, owner_id, type)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, channel.Name, channel.OwnerID, channel.ChannelType)
	} else {
		row = r.db.QueryRowContext(ctx, query, channel.Name, channel.OwnerID, channel.ChannelType)
	}

	err := row.Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (r *channelsRepo) FindDMChannelByUserIDs(ctx context.Context, userID, targetUserID uuid.UUID) (*Channel, error) {
	query := `
		SELECT c.id, c.name, c.owner_id, c.type, c.created_at, c.updated_at
		FROM channels c
		JOIN channel_members cm1 ON c.id = cm1.channel_id
		JOIN channel_members cm2 ON c.id = cm2.channel_id
		WHERE c.type = 'dm'
		AND cm1.user_id = $1
		AND cm2.user_id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx, query, userID, targetUserID)
	channel := &Channel{}
	err := row.Scan(&channel.ID, &channel.Name, &channel.OwnerID, &channel.ChannelType, &channel.CreatedAt, &channel.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return channel, nil
}

func (r *channelsRepo) SaveDMChannel(ctx context.Context, userId, targetUserID uuid.UUID) (*Channel, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	dmName := fmt.Sprintf("dm_%s_%s", userId.String(), targetUserID.String())

	newChannel := &Channel{
		Name:        dmName,
		OwnerID:     userId,
		ChannelType: ChannelTypeDM,
	}

	savedChannel, err := r.SaveChannel(ctx, newChannel, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	channelID, err := uuid.Parse(savedChannel.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	if _, err := r.AddUserToChannel(ctx, channelID, userId, tx); err != nil {
		return nil, fmt.Errorf("failed to add owner to channel: %w", err)
	}

	if _, err := r.AddUserToChannel(ctx, channelID, targetUserID, tx); err != nil {
		return nil, fmt.Errorf("failed to add target user to channel: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return savedChannel, nil
}

func (r *channelsRepo) AddUserToChannel(ctx context.Context, channelID, userID uuid.UUID, tx *sql.Tx) (uuid.UUID, error) {
	query := `
		INSERT INTO channel_members (channel_id, user_id)
		VALUES ($1, $2)
	`
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, channelID, userID)
	} else {
		_, err = r.db.ExecContext(ctx, query, channelID, userID)
	}
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *channelsRepo) SaveGroupChannel(ctx context.Context, ownerUserID uuid.UUID, name string, channelMemberIDs []uuid.UUID) (*Channel, []uuid.UUID, error) {
	log.Printf("saving group channel %s with owner %s and users %v", name, ownerUserID, channelMemberIDs)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	newChannel := &Channel{
		Name:        name,
		OwnerID:     ownerUserID,
		ChannelType: ChannelTypeGroup,
	}

	savedChannel, err := r.SaveChannel(ctx, newChannel, tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save channel: %w", err)
	}

	channelID, err := uuid.Parse(savedChannel.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	savedOwnerID, err := r.AddUserToChannel(ctx, channelID, ownerUserID, tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add owner to channel: %w", err)
	}

	memberUserIDs := []uuid.UUID{savedOwnerID}

	for _, userID := range channelMemberIDs {
		log.Printf("adding user %s to channel %s", userID, channelID)
		savedUserID, err := r.AddUserToChannel(ctx, channelID, userID, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add user to channel: %w", err)
		}
		memberUserIDs = append(memberUserIDs, savedUserID)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("channel members: %v", memberUserIDs)

	return savedChannel, memberUserIDs, nil
}

func (r *channelsRepo) FindAllChannelsByUserID(ctx context.Context, userID uuid.UUID) ([]*ChannelWithMembers, error) {
	query := `
		SELECT 
			c.id, c.name, c.owner_id, c.type, c.created_at, c.updated_at, 
			cm.user_id
		FROM channels c
		JOIN channel_members cm ON c.id = cm.channel_id
		WHERE c.id IN (
			SELECT c2.id
			FROM channels c2
			JOIN channel_members cm2 ON c2.id = cm2.channel_id
			WHERE cm2.user_id = $1
		)
		ORDER BY c.created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	channelMap := map[string]*ChannelWithMembers{}

	for rows.Next() {
		var (
			channelID    string
			name         string
			ownerID      uuid.UUID
			channelType  ChannelType
			createdAt    time.Time
			updatedAt    time.Time
			memberUserID uuid.UUID
		)

		err := rows.Scan(&channelID, &name, &ownerID, &channelType, &createdAt, &updatedAt, &memberUserID)
		if err != nil {
			return nil, err
		}

		if _, exists := channelMap[channelID]; !exists {
			channelMap[channelID] = &ChannelWithMembers{
				Channel: &Channel{
					ID:          channelID,
					Name:        name,
					OwnerID:     ownerID,
					ChannelType: channelType,
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				},
				Members: []uuid.UUID{},
			}
		}

		channelMap[channelID].Members = append(channelMap[channelID].Members, memberUserID)
	}

	result := make([]*ChannelWithMembers, 0, len(channelMap))
	for _, v := range channelMap {
		result = append(result, v)
	}

	return result, nil
}

func (r *channelsRepo) FindDMChannelByID(ctx context.Context, channelID uuid.UUID) (*Channel, error) {
	query := `
		SELECT id, name, owner_id, type, created_at, updated_at
		FROM channels
		WHERE id = $1 AND type = 'dm'
	`
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx, query, channelID)
	channel := &Channel{}
	err := row.Scan(&channel.ID, &channel.Name, &channel.OwnerID, &channel.ChannelType, &channel.CreatedAt, &channel.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewNotFoundError("DM channel not found")
		}
		return nil, err
	}

	return channel, nil
}
