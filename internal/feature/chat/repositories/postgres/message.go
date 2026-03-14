package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zercle/zercle-go-template/internal/feature/chat/domain"
	"github.com/zercle/zercle-go-template/internal/feature/chat/ports"
)

// MessageRepository implements ports.MessageRepository for PostgreSQL.
type MessageRepository struct {
	db *pgxpool.Pool
}

// NewMessageRepository creates a new PostgreSQL message repository.
func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message in the database.
func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, content, message_type, reply_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		message.ID,
		message.RoomID,
		message.SenderID,
		message.Content,
		message.MessageType,
		message.ReplyTo,
		message.CreatedAt,
		message.UpdatedAt,
	)
	return err
}

// FindByID finds a message by ID.
func (r *MessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	query := `
		SELECT m.id, m.room_id, m.sender_id, u.username, m.content, m.message_type, m.reply_to, m.created_at, m.updated_at, m.deleted_at
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`
	var message domain.Message
	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.RoomID,
		&message.SenderID,
		&message.SenderUsername,
		&message.Content,
		&message.MessageType,
		&message.ReplyTo,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMessageNotFound
	}
	return &message, err
}

// FindByRoomID finds messages for a room with pagination.
func (r *MessageRepository) FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error) {
	var query string
	var rows pgx.Rows
	var err error

	if before != nil {
		query = `
			SELECT m.id, m.room_id, m.sender_id, u.username, m.content, m.message_type, m.reply_to, m.created_at, m.updated_at, m.deleted_at
			FROM messages m
			LEFT JOIN users u ON m.sender_id = u.id
			WHERE m.room_id = $1 AND m.deleted_at IS NULL AND m.created_at < (
				SELECT created_at FROM messages WHERE id = $2
			)
			ORDER BY m.created_at DESC
			LIMIT $3 OFFSET $4
		`
		rows, err = r.db.Query(ctx, query, roomID, before, limit, offset)
	} else {
		query = `
			SELECT m.id, m.room_id, m.sender_id, u.username, m.content, m.message_type, m.reply_to, m.created_at, m.updated_at, m.deleted_at
			FROM messages m
			LEFT JOIN users u ON m.sender_id = u.id
			WHERE m.room_id = $1 AND m.deleted_at IS NULL
			ORDER BY m.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(ctx, query, roomID, limit, offset)
	}

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &m.SenderUsername, &m.Content, &m.MessageType, &m.ReplyTo, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, false, err
		}
		messages = append(messages, &m)
	}

	hasMore := len(messages) == limit
	return messages, hasMore, nil
}

// Delete soft-deletes a message.
func (r *MessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE messages SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

var _ ports.MessageRepository = (*MessageRepository)(nil)
