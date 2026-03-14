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

// RoomRepository implements ports.RoomRepository for PostgreSQL.
type RoomRepository struct {
	db *pgxpool.Pool
}

// NewRoomRepository creates a new PostgreSQL room repository.
func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create creates a new room in the database.
func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, name, description, type, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
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
	_, err = r.db.Exec(ctx, memberQuery, room.ID, room.OwnerID, "owner", room.CreatedAt)
	return err
}

// FindByID finds a room by ID.
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
	err := r.db.QueryRow(ctx, query, id).Scan(
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
		return nil, domain.ErrRoomNotFound
	}
	return &room, err
}

// FindByUserID finds rooms for a user.
func (r *RoomRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM room_members rm
		JOIN rooms r ON rm.room_id = r.id
		WHERE rm.user_id = $1 AND r.deleted_at IS NULL
	`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT r.id, r.name, r.description, r.type, r.owner_id,
			   COUNT(rm.user_id) as member_count, r.created_at, r.updated_at, r.deleted_at
		FROM rooms r
		LEFT JOIN room_members rm ON r.id = rm.room_id
		JOIN room_members my_rm ON r.id = my_rm.room_id AND my_rm.user_id = $1
		WHERE r.deleted_at IS NULL
		GROUP BY r.id
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(
			&room.ID, &room.Name, &room.Description, &room.Type, &room.OwnerID,
			&room.MemberCount, &room.CreatedAt, &room.UpdatedAt, &room.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, &room)
	}
	return rooms, total, nil
}

// Update updates an existing room.
func (r *RoomRepository) Update(ctx context.Context, room *domain.Room) error {
	query := `
		UPDATE rooms
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, room.ID, room.Name, room.Description, room.UpdatedAt)
	return err
}

// Delete soft-deletes a room.
func (r *RoomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE rooms SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// AddMember adds a member to a room.
func (r *RoomRepository) AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (room_id, user_id) DO UPDATE SET role = $3
	`
	_, err := r.db.Exec(ctx, query, roomID, userID, role)
	return err
}

// RemoveMember removes a member from a room.
func (r *RoomRepository) RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, roomID, userID)
	return err
}

// GetMembers returns all members of a room.
func (r *RoomRepository) GetMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error) {
	query := `
		SELECT rm.room_id, rm.user_id, u.username, u.display_name, u.avatar_url, rm.role, rm.joined_at
		FROM room_members rm
		JOIN users u ON rm.user_id = u.id
		WHERE rm.room_id = $1
	`
	rows, err := r.db.Query(ctx, query, roomID)
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

// IsMember checks if a user is a member of a room.
func (r *RoomRepository) IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, roomID, userID).Scan(&exists)
	return exists, err
}

var _ ports.RoomRepository = (*RoomRepository)(nil)
