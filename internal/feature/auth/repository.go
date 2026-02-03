package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	sqlcauth "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/auth"
	sqlccredential "github.com/zercle/zercle-go-template/internal/infrastructure/database/sqlc/credential"
)

// CredentialRepository defines the interface for credential storage.
type CredentialRepository interface {
	Create(ctx context.Context, credential *Credential) (*Credential, error)
	GetByUserID(ctx context.Context, userID string) (*Credential, error)
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

// RefreshTokenRepository defines the interface for refresh token storage.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) (*RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type postgresCredentialRepository struct {
	queries *sqlccredential.Queries
}

// NewCredentialRepository creates a new Postgres credential repository.
func NewCredentialRepository(db sqlccredential.DBTX) CredentialRepository {
	return &postgresCredentialRepository{
		queries: sqlccredential.New(db),
	}
}

func mapCredentialPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrEmailAlreadyExists
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidCredentials
	}

	return err
}

func (r *postgresCredentialRepository) Create(ctx context.Context, credential *Credential) (*Credential, error) {
	params := &sqlccredential.CreateCredentialParams{
		ID:           credential.ID,
		UserID:       credential.UserID,
		PasswordHash: credential.PasswordHash,
		CreatedAt:    toPgTimestamptz(credential.CreatedAt),
		UpdatedAt:    toPgTimestamptz(credential.UpdatedAt),
	}
	result, err := r.queries.CreateCredential(ctx, params)
	if err != nil {
		return nil, mapCredentialPgError(err)
	}
	return sqlcCredentialToDomain(result)
}

func (r *postgresCredentialRepository) GetByUserID(ctx context.Context, userID string) (*Credential, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	result, err := r.queries.GetCredentialByUserID(ctx, uid)
	if err != nil {
		return nil, mapCredentialPgError(err)
	}
	return sqlcCredentialToDomain(result)
}

func (r *postgresCredentialRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	params := &sqlccredential.UpdateCredentialPasswordParams{
		UserID:       uid,
		PasswordHash: passwordHash,
		UpdatedAt:    toPgTimestamptz(time.Now()),
	}
	_, err = r.queries.UpdateCredentialPassword(ctx, params)
	return mapCredentialPgError(err)
}

func (r *postgresCredentialRepository) DeleteByUserID(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	err = r.queries.DeleteCredentialByUserID(ctx, uid)
	return mapCredentialPgError(err)
}

func sqlcCredentialToDomain(m *sqlccredential.UserCredential) (*Credential, error) {
	createdAt := time.Time{}
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt.Time
	}

	return &Credential{
		ID:           m.ID,
		UserID:       m.UserID,
		PasswordHash: m.PasswordHash,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

type postgresRefreshTokenRepository struct {
	queries *sqlcauth.Queries
}

// NewRefreshTokenRepository creates a new Postgres refresh token repository.
func NewRefreshTokenRepository(db sqlcauth.DBTX) RefreshTokenRepository {
	return &postgresRefreshTokenRepository{
		queries: sqlcauth.New(db),
	}
}

func (r *postgresRefreshTokenRepository) Create(ctx context.Context, token *RefreshToken) (*RefreshToken, error) {
	params := &sqlcauth.CreateRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: toPgTimestamptz(token.ExpiresAt),
		CreatedAt: toPgTimestamptz(token.CreatedAt),
	}
	if token.RevokedAt != nil {
		params.RevokedAt = toPgTimestamptz(*token.RevokedAt)
	}
	result, err := r.queries.CreateRefreshToken(ctx, params)
	if err != nil {
		return nil, err
	}
	return sqlcRefreshTokenToDomain(result)
}

func (r *postgresRefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	result, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	return sqlcRefreshTokenToDomain(result)
}

func (r *postgresRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.queries.RevokeRefreshToken(ctx, tokenHash)
	return err
}

func (r *postgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	err = r.queries.RevokeAllUserRefreshTokens(ctx, uid)
	return err
}

func sqlcRefreshTokenToDomain(m *sqlcauth.RefreshToken) (*RefreshToken, error) {
	expiresAt := time.Time{}
	if m.ExpiresAt.Valid {
		expiresAt = m.ExpiresAt.Time
	}
	var revokedAt *time.Time
	if m.RevokedAt.Valid {
		revokedAt = &m.RevokedAt.Time
	}
	createdAt := time.Time{}
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}

	return &RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: expiresAt,
		RevokedAt: revokedAt,
		CreatedAt: createdAt,
	}, nil
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
