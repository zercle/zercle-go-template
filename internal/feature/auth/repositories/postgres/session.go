package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zercle/zercle-go-template/internal/feature/auth/domain"
	"github.com/zercle/zercle-go-template/internal/feature/auth/ports"
)

// SessionRepository implements ports.SessionRepository for PostgreSQL.
type SessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository creates a new PostgreSQL session repository.
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session in the database.
func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query,
		session.UserID,
		session.Token,
		session.ExpiresAt,
	)
	return err
}

// FindByToken finds a session by token.
func (r *SessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT user_id, token, expires_at
		FROM sessions
		WHERE token = $1
	`
	var session domain.Session
	err := r.db.QueryRow(ctx, query, token).Scan(
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTokenInvalid
	}
	return &session, err
}

// Delete deletes a session by token.
func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Exec(ctx, query, token)
	return err
}

// DeleteByUserID deletes all sessions for a user.
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

var _ ports.SessionRepository = (*SessionRepository)(nil)
