package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zercle/zercle-go-template/internal/features/auth/domain"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// SessionRepository handles session/token persistence in PostgreSQL.
type SessionRepository struct {
	db *DB
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create stores a new refresh token session.
func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		uuidgen.New(),
		session.UserID,
		session.Token,
		session.ExpiresAt,
		time.Now(),
	)
	return err
}

// FindByToken retrieves a session by its token.
func (r *SessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT user_id, token, expires_at
		FROM refresh_tokens
		WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`
	var session domain.Session
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrTokenInvalid
	}
	return &session, err
}

// Delete revokes a session by token.
func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1`
	_, err := r.db.Pool.Exec(ctx, query, token)
	return err
}

// DeleteByUserID revokes all sessions for a user.
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	return err
}
