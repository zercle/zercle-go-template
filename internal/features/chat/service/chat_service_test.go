package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/features/chat/domain"

	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

type mockRoomRepo struct {
	rooms   map[uuid.UUID]*domain.Room
	members map[string]bool
}

func (m *mockRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	room, ok := m.rooms[id]
	if !ok {
		return nil, apperrors.ErrRoomNotFound
	}
	return room, nil
}

func (m *mockRoomRepo) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error) {
	var result []*domain.Room
	for _, room := range m.rooms {
		key := roomKey(room.ID, userID)
		if m.members[key] {
			result = append(result, room)
		}
	}
	total := len(result)
	if offset >= len(result) {
		return []*domain.Room{}, total, nil
	}
	end := min(offset+limit, len(result))
	return result[offset:end], total, nil
}

func (m *mockRoomRepo) GetMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error) {
	var members []*domain.RoomMember
	for key := range m.members {
		if key[:len(key)-len("-member")-len("00000000-0000-0000-0000-000000000000")+1] == roomID.String()[:len(roomID.String())-1] {
			var roomMember domain.RoomMember
			roomMember.RoomID = roomID
			members = append(members, &roomMember)
		}
	}
	return members, nil
}

func (m *mockRoomRepo) IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	return m.members[roomKey(roomID, userID)], nil
}

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.Room) error {
	m.rooms[room.ID] = room
	return nil
}

func (m *mockRoomRepo) Update(ctx context.Context, room *domain.Room) error {
	m.rooms[room.ID] = room
	return nil
}

func (m *mockRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.rooms, id)
	return nil
}

func (m *mockRoomRepo) AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) error {
	m.members[roomKey(roomID, userID)] = true
	if room, ok := m.rooms[roomID]; ok {
		room.MemberCount++
	}
	return nil
}

func (m *mockRoomRepo) RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error {
	delete(m.members, roomKey(roomID, userID))
	if room, ok := m.rooms[roomID]; ok {
		room.MemberCount--
	}
	return nil
}

func roomKey(roomID, userID uuid.UUID) string {
	return roomID.String() + ":" + userID.String()
}

type mockMessageRepo struct {
	messages map[uuid.UUID]*domain.Message
}

func (m *mockMessageRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	msg, ok := m.messages[id]
	if !ok {
		return nil, apperrors.ErrMessageNotFound
	}
	return msg, nil
}

func (m *mockMessageRepo) FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error) {
	var result []*domain.Message
	for _, msg := range m.messages {
		if msg.RoomID == roomID {
			result = append(result, msg)
		}
	}
	if offset >= len(result) {
		return []*domain.Message{}, false, nil
	}
	return result[offset:min(offset+limit, len(result))], false, nil
}

func (m *mockMessageRepo) Create(ctx context.Context, message *domain.Message) error {
	m.messages[message.ID] = message
	return nil
}

func (m *mockMessageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.messages, id)
	return nil
}

func newMockRoomRepo() *mockRoomRepo {
	return &mockRoomRepo{
		rooms:   make(map[uuid.UUID]*domain.Room),
		members: make(map[string]bool),
	}
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: make(map[uuid.UUID]*domain.Message),
	}
}

func TestChatService_CreateRoom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateRoomInput
		wantErr bool
		errType error
	}{
		{
			name: "success with owner added as member",
			input: CreateRoomInput{
				Name:        "Test Room",
				Description: "Test Description",
				Type:        domain.RoomTypePublic,
				OwnerID:     uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "validation error for empty name",
			input: CreateRoomInput{
				Name:        "",
				Description: "Test Description",
				Type:        domain.RoomTypePublic,
				OwnerID:     uuid.New(),
			},
			wantErr: true,
			errType: apperrors.ErrRoomNameRequired,
		},
		{
			name: "success with additional members",
			input: CreateRoomInput{
				Name:        "Room With Members",
				Description: "Members test",
				Type:        domain.RoomTypePrivate,
				OwnerID:     uuid.New(),
				MemberIDs:   []uuid.UUID{uuid.New(), uuid.New()},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			roomRepo := newMockRoomRepo()
			messageRepo := newMockMessageRepo()
			svc := NewChatService(roomRepo, messageRepo, nil)

			room, err := svc.CreateRoom(context.Background(), tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errType != nil && !errors.Is(err, tc.errType) {
					t.Fatalf("expected error %v, got %v", tc.errType, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if room.Name != tc.input.Name {
				t.Errorf("expected name %s, got %s", tc.input.Name, room.Name)
			}
			if room.Description != tc.input.Description {
				t.Errorf("expected description %s, got %s", tc.input.Description, room.Description)
			}
			if room.Type != tc.input.Type {
				t.Errorf("expected type %s, got %s", tc.input.Type, room.Type)
			}
			if room.OwnerID != tc.input.OwnerID {
				t.Errorf("expected owner %s, got %s", tc.input.OwnerID, room.OwnerID)
			}
			isOwnerMember, _ := roomRepo.IsMember(context.Background(), room.ID, tc.input.OwnerID)
			if !isOwnerMember {
				t.Error("expected owner to be a member")
			}
		})
	}
}

func TestChatService_GetRoom(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "Description", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room

		got, err := svc.GetRoom(context.Background(), room.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != room.ID {
			t.Errorf("expected room %s, got %s", room.ID, got.ID)
		}
	})

	t.Run("room not found", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		_, err := svc.GetRoom(context.Background(), uuid.New())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperrors.ErrRoomNotFound) {
			t.Fatalf("expected ErrRoomNotFound, got %v", err)
		}
	})
}

func TestChatService_ListRooms(t *testing.T) {
	t.Parallel()

	t.Run("default pagination", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		userID := uuid.New()
		for range 5 {
			room := domain.NewRoom("Room", "", domain.RoomTypePublic, userID)
			roomRepo.rooms[room.ID] = room
			roomRepo.members[roomKey(room.ID, userID)] = true
		}

		rooms, total, err := svc.ListRooms(context.Background(), userID, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(rooms) != 5 {
			t.Errorf("expected 5 rooms, got %d", len(rooms))
		}
	})

	t.Run("custom limit", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		userID := uuid.New()
		for range 10 {
			room := domain.NewRoom("Room", "", domain.RoomTypePublic, userID)
			roomRepo.rooms[room.ID] = room
			roomRepo.members[roomKey(room.ID, userID)] = true
		}

		rooms, total, err := svc.ListRooms(context.Background(), userID, 5, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 10 {
			t.Errorf("expected total 10, got %d", total)
		}
		if len(rooms) != 5 {
			t.Errorf("expected 5 rooms, got %d", len(rooms))
		}
	})

	t.Run("max limit capped", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		userID := uuid.New()
		for range 150 {
			room := domain.NewRoom("Room", "", domain.RoomTypePublic, userID)
			roomRepo.rooms[room.ID] = room
			roomRepo.members[roomKey(room.ID, userID)] = true
		}

		rooms, _, err := svc.ListRooms(context.Background(), userID, 200, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rooms) != MaxRoomPageSize {
			t.Errorf("expected %d rooms (max), got %d", MaxRoomPageSize, len(rooms))
		}
	})
}

func TestChatService_JoinRoom(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room
		roomRepo.members[roomKey(room.ID, ownerID)] = true

		userID := uuid.New()
		err := svc.JoinRoom(context.Background(), room.ID, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		isMember, _ := roomRepo.IsMember(context.Background(), room.ID, userID)
		if !isMember {
			t.Error("expected user to be a member")
		}
	})

	t.Run("already joined", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room
		roomRepo.members[roomKey(room.ID, ownerID)] = true

		err := svc.JoinRoom(context.Background(), room.ID, ownerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperrors.ErrAlreadyJoined) {
			t.Fatalf("expected ErrAlreadyJoined, got %v", err)
		}
	})
}

func TestChatService_LeaveRoom(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room
		roomRepo.members[roomKey(room.ID, ownerID)] = true

		err := svc.LeaveRoom(context.Background(), room.ID, ownerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		isMember, _ := roomRepo.IsMember(context.Background(), room.ID, ownerID)
		if isMember {
			t.Error("expected user to not be a member")
		}
	})

	t.Run("not a member", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room
		roomRepo.members[roomKey(room.ID, ownerID)] = true

		err := svc.LeaveRoom(context.Background(), room.ID, uuid.New())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperrors.ErrNotMember) {
			t.Fatalf("expected ErrNotMember, got %v", err)
		}
	})
}

func TestChatService_SendMessage(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room
		roomRepo.members[roomKey(room.ID, ownerID)] = true

		input := SendMessageInput{
			RoomID:      room.ID,
			SenderID:    ownerID,
			Content:     "Hello, world!",
			MessageType: MessageTypeText,
		}

		msg, err := svc.SendMessage(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Content != input.Content {
			t.Errorf("expected content %s, got %s", input.Content, msg.Content)
		}
	})

	t.Run("not a member", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room
		roomRepo.members[roomKey(room.ID, ownerID)] = true

		input := SendMessageInput{
			RoomID:      room.ID,
			SenderID:    uuid.New(),
			Content:     "Hello, world!",
			MessageType: MessageTypeText,
		}

		_, err := svc.SendMessage(context.Background(), input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, apperrors.ErrNotMember) {
			t.Fatalf("expected ErrNotMember, got %v", err)
		}
	})
}

func TestChatService_GetMessageHistory(t *testing.T) {
	t.Parallel()

	t.Run("default pagination", func(t *testing.T) {
		t.Parallel()
		roomRepo := newMockRoomRepo()
		messageRepo := newMockMessageRepo()
		svc := NewChatService(roomRepo, messageRepo, nil)

		ownerID := uuid.New()
		room := domain.NewRoom("Test Room", "", domain.RoomTypePublic, ownerID)
		roomRepo.rooms[room.ID] = room

		for range 60 {
			msg := &domain.Message{
				ID:          uuid.New(),
				RoomID:      room.ID,
				SenderID:    ownerID,
				Content:     "Message",
				MessageType: MessageTypeText,
				CreatedAt:   time.Now(),
			}
			messageRepo.messages[msg.ID] = msg
		}

		messages, _, err := svc.GetMessageHistory(context.Background(), room.ID, 0, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != DefaultMessagePageSize {
			t.Errorf("expected %d messages, got %d", DefaultMessagePageSize, len(messages))
		}
	})
}
