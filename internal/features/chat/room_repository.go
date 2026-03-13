package chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	apperrors "github.com/zercle/zercle-go-template/internal/errors"
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/postgres"
)

type RoomRepository struct {
	db *postgres.DB
}

func NewRoomRepository(db *postgres.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, name, description, type, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		room.ID,
		room.Name,
		room.Description,
		room.Type,
		room.OwnerID,
		room.CreatedAt,
		room.UpdatedAt,
	)
	if err != nil {
		return err
	}

	memberQuery := `
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = r.db.Pool.Exec(ctx, memberQuery, room.ID, room.OwnerID, "owner", room.CreatedAt)
	return err
}

func (r *RoomRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.type, r.owner_id, 
			   COUNT(rm.user_id) as member_count, r.created_at, r.updated_at, r.deleted_at
		FROM rooms r
		LEFT JOIN room_members rm ON r.id = rm.room_id
		WHERE r.id = $1 AND r.deleted_at IS NULL
		GROUP BY r.id
	`
	var room domain.Room
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.Type,
		&room.OwnerID,
		&room.MemberCount,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrRoomNotFound
	}
	return &room, err
}

func (r *RoomRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM room_members rm
		JOIN rooms r ON rm.room_id = r.id
		WHERE rm.user_id = $1 AND r.deleted_at IS NULL
	`
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT r.id, r.name, r.description, r.type, r.owner_id,
			   COUNT(rm2.user_id) as member_count, r.created_at, r.updated_at, r.deleted_at
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id AND rm.user_id = $1
		LEFT JOIN room_members rm2 ON r.id = rm2.room_id
		WHERE r.deleted_at IS NULL
		GROUP BY r.id
		ORDER BY r.updated_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.Type,
			&room.OwnerID,
			&room.MemberCount,
			&room.CreatedAt,
			&room.UpdatedAt,
			&room.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, total, nil
}

func (r *RoomRepository) Update(ctx context.Context, room *domain.Room) error {
	query := `
		UPDATE rooms
		SET name = $2, description = $3, type = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Pool.Exec(ctx, query,
		room.ID,
		room.Name,
		room.Description,
		room.Type,
		room.UpdatedAt,
	)
	return err
}

func (r *RoomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE rooms SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *RoomRepository) AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (room_id, user_id) DO UPDATE SET role = $3
	`
	_, err := r.db.Pool.Exec(ctx, query, roomID, userID, role)
	return err
}

func (r *RoomRepository) RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, roomID, userID)
	return err
}

func (r *RoomRepository) GetMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error) {
	query := `
		SELECT rm.room_id, rm.user_id, u.username, u.display_name, u.avatar_url, rm.role, rm.joined_at
		FROM room_members rm
		JOIN users u ON rm.user_id = u.id
		WHERE rm.room_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.RoomMember
	for rows.Next() {
		var m domain.RoomMember
		if err := rows.Scan(&m.RoomID, &m.UserID, &m.Username, &m.DisplayName, &m.AvatarURL, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *RoomRepository) IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, roomID, userID).Scan(&exists)
	return exists, err
}
