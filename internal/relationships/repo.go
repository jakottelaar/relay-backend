package relationships

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type RelationshipsRepo interface {
	SaveRelationship(ctx context.Context, current_user_id, target_user_id uuid.UUID) (*Relationship, error)
	FindRelationshipByUserIDAndOtherUserID(ctx context.Context, userID, otherUserID uuid.UUID) (*Relationship, error)
	UpdateRelationshipStatus(ctx context.Context, userID, otherUserID uuid.UUID, status RelationshipStatus) (*Relationship, error)
}

type relationshipsRepo struct {
	db *sql.DB
}

func NewRelationshipsRepo(db *sql.DB) RelationshipsRepo {
	return &relationshipsRepo{db: db}
}

func (r *relationshipsRepo) SaveRelationship(ctx context.Context, current_user_id, target_user_id uuid.UUID) (*Relationship, error) {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			// Rollback the transaction if there's an error
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			return
		}
		// Commit the transaction if no errors
		if err = tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
		}
	}()

	// Prepare the queries for creating symmetric relationships
	outgoingQuery := `
        INSERT INTO relationships (user_id, other_user_id, relationship_status) 
        VALUES ($1, $2, 'outgoing') 
        RETURNING id, created_at, updated_at
    `
	incomingQuery := `
        INSERT INTO relationships (user_id, other_user_id, relationship_status) 
        VALUES ($1, $2, 'incoming') 
        RETURNING id, created_at, updated_at
    `

	// Create the outgoing relationship
	outgoingRelationship := &Relationship{
		UserID:             current_user_id,
		OtherUserID:        target_user_id,
		RelationshipStatus: RelationshipStatusOutgoing,
	}

	// Execute outgoing relationship insertion
	err = tx.QueryRowContext(ctx, outgoingQuery, current_user_id, target_user_id).Scan(
		&outgoingRelationship.ID,
		&outgoingRelationship.CreatedAt,
		&outgoingRelationship.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create outgoing relationship: %w", err)
	}

	// Create the incoming relationship
	incomingRelationship := &Relationship{
		UserID:             target_user_id,
		OtherUserID:        current_user_id,
		RelationshipStatus: RelationshipStatusIncoming,
	}

	// Execute incoming relationship insertion
	err = tx.QueryRowContext(ctx, incomingQuery, target_user_id, current_user_id).Scan(
		&incomingRelationship.ID,
		&incomingRelationship.CreatedAt,
		&incomingRelationship.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create incoming relationship: %w", err)
	}

	return outgoingRelationship, nil
}

func (r *relationshipsRepo) FindRelationshipByUserIDAndOtherUserID(ctx context.Context, userID, otherUserID uuid.UUID) (*Relationship, error) {
	query := `SELECT id, user_id, other_user_id, relationship_status, created_at, updated_at 
              FROM relationships
              WHERE 
                (user_id = $1 AND other_user_id = $2) 
                OR (user_id = $2 AND other_user_id = $1)
              ORDER BY 
                CASE 
                    WHEN user_id = $1 AND other_user_id = $2 THEN 1
                    WHEN user_id = $2 AND other_user_id = $1 THEN 2
                END`

	var relationship Relationship
	err := r.db.QueryRowContext(ctx, query, userID, otherUserID).Scan(
		&relationship.ID,
		&relationship.UserID,
		&relationship.OtherUserID,
		&relationship.RelationshipStatus,
		&relationship.CreatedAt,
		&relationship.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &relationship, nil
}

func (r *relationshipsRepo) UpdateRelationshipStatus(ctx context.Context, userID, otherUserID uuid.UUID, status RelationshipStatus) (*Relationship, error) {
	query := `UPDATE relationships SET relationship_status = 'friend'
	WHERE user_id = $1 AND other_user_id = $2
	RETURNING id, user_id, other_user_id, relationship_status, created_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	row := r.db.QueryRowContext(ctx, query, userID, otherUserID)

	var relationship Relationship
	err := row.Scan(&relationship.ID, &relationship.UserID, &relationship.OtherUserID, &relationship.RelationshipStatus, &relationship.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &relationship, nil
}
