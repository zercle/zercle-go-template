package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/zercle/zercle-go-template/api/pb"
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/features/chat/service"
)

// ChatServer implements the gRPC chat service server.
type ChatServer struct {
	pb.UnimplementedChatServiceServer
	chatService *service.ChatService
}

// NewChatServer creates a new ChatServer with the given chat service.
func NewChatServer(chatService *service.ChatService) *ChatServer {
	return &ChatServer{chatService: chatService}
}

// CreateRoom creates a new chat room.
func (s *ChatServer) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Room, error) {
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, fmt.Errorf("failed to parse owner ID: %w", err)
	}

	memberIDs := make([]uuid.UUID, 0)
	for _, idStr := range req.GetMemberIds() {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		memberIDs = append(memberIDs, id)
	}

	input := service.CreateRoomInput{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Type:        req.GetType(),
		OwnerID:     ownerID,
		MemberIDs:   memberIDs,
	}

	room, err := s.chatService.CreateRoom(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return toProtoRoom(room), nil
}

// GetRoom retrieves a room by ID.
func (s *ChatServer) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.Room, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	room, err := s.chatService.GetRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return toProtoRoom(room), nil
}

// UpdateRoom updates room name and description.
func (s *ChatServer) UpdateRoom(ctx context.Context, req *pb.UpdateRoomRequest) (*pb.Room, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	room, err := s.chatService.UpdateRoom(ctx, roomID, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	return toProtoRoom(room), nil
}

// DeleteRoom deletes a room by ID.
func (s *ChatServer) DeleteRoom(ctx context.Context, req *pb.DeleteRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	if err := s.chatService.DeleteRoom(ctx, roomID); err != nil {
		return nil, fmt.Errorf("failed to delete room: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// ListRooms lists rooms for a user with pagination.
func (s *ChatServer) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	rooms, total, err := s.chatService.ListRooms(ctx, userID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	protoRooms := make([]*pb.Room, len(rooms))
	for i, room := range rooms {
		protoRooms[i] = toProtoRoom(room)
	}

	return &pb.ListRoomsResponse{
		Rooms: protoRooms,
		Total: int32(total), //nolint:gosec // safe: total from DB count, fits in int32
	}, nil
}

// JoinRoom adds a user to a room.
func (s *ChatServer) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	if err := s.chatService.JoinRoom(ctx, roomID, userID); err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// LeaveRoom removes a user from a room.
func (s *ChatServer) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	if err := s.chatService.LeaveRoom(ctx, roomID, userID); err != nil {
		return nil, fmt.Errorf("failed to leave room: %w", err)
	}

	return &emptypb.Empty{}, nil
}

// GetRoomMembers retrieves all members of a room.
func (s *ChatServer) GetRoomMembers(ctx context.Context, req *pb.GetRoomMembersRequest) (*pb.GetRoomMembersResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	members, err := s.chatService.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}

	protoMembers := make([]*pb.RoomMember, len(members))
	for i, m := range members {
		protoMembers[i] = &pb.RoomMember{
			UserId:      m.UserID.String(),
			Username:    m.Username,
			DisplayName: m.DisplayName,
			AvatarUrl:   m.AvatarURL,
			Role:        m.Role,
			JoinedAt:    timestamppb.New(m.JoinedAt),
		}
	}

	return &pb.GetRoomMembersResponse{Members: protoMembers}, nil
}

// SendMessage sends a message to a room.
func (s *ChatServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	senderID, err := uuid.Parse(req.SenderId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sender ID: %w", err)
	}

	var replyTo *uuid.UUID
	if req.ReplyTo != "" {
		id, err := uuid.Parse(req.ReplyTo)
		if err == nil {
			replyTo = &id
		}
	}

	input := service.SendMessageInput{
		RoomID:      roomID,
		SenderID:    senderID,
		Content:     req.Content,
		MessageType: req.MessageType,
		ReplyTo:     replyTo,
	}

	message, err := s.chatService.SendMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return toProtoMessage(message), nil
}

// GetMessageHistory retrieves paginated message history for a room.
func (s *ChatServer) GetMessageHistory(ctx context.Context, req *pb.GetMessageHistoryRequest) (*pb.GetMessageHistoryResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse room ID: %w", err)
	}

	var before *uuid.UUID
	if req.Before != "" {
		id, err := uuid.Parse(req.Before)
		if err == nil {
			before = &id
		}
	}

	messages, hasMore, err := s.chatService.GetMessageHistory(ctx, roomID, int(req.Limit), int(req.Offset), before)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	protoMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		protoMessages[i] = toProtoMessage(msg)
	}

	return &pb.GetMessageHistoryResponse{
		Messages: protoMessages,
		HasMore:  hasMore,
	}, nil
}

// ChatStream handles bidirectional streaming for real-time chat.
func (s *ChatServer) ChatStream(stream pb.ChatService_ChatStreamServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to receive stream message: %w", err)
		}

		switch payload := req.Payload.(type) {
		case *pb.ChatStreamRequest_Join:
			roomID, _ := uuid.Parse(payload.Join.RoomId)
			userID, _ := uuid.Parse(payload.Join.UserId)
			if err := s.chatService.JoinRoom(stream.Context(), roomID, userID); err != nil {
				return fmt.Errorf("failed to join room via stream: %w", err)
			}
		case *pb.ChatStreamRequest_Leave:
			roomID, _ := uuid.Parse(payload.Leave.RoomId)
			userID, _ := uuid.Parse(payload.Leave.UserId)
			if err := s.chatService.LeaveRoom(stream.Context(), roomID, userID); err != nil {
				return fmt.Errorf("failed to leave room via stream: %w", err)
			}
		case *pb.ChatStreamRequest_Message:
			msg := payload.Message
			roomID, _ := uuid.Parse(msg.RoomId)
			senderID, _ := uuid.Parse(msg.SenderId)
			input := service.SendMessageInput{
				RoomID:      roomID,
				SenderID:    senderID,
				Content:     msg.Content,
				MessageType: msg.MessageType,
			}
			sentMsg, err := s.chatService.SendMessage(stream.Context(), input)
			if err != nil {
				return fmt.Errorf("failed to send message via stream: %w", err)
			}
			if err := stream.Send(&pb.ChatStreamResponse{
				Payload: &pb.ChatStreamResponse_Message{
					Message: toProtoMessage(sentMsg),
				},
			}); err != nil {
				return fmt.Errorf("failed to send stream response: %w", err)
			}
		}
	}
}

// SetPresence updates user presence status in a room.
func (s *ChatServer) SetPresence(ctx context.Context, req *pb.SetPresenceRequest) (*emptypb.Empty, error) {
	_ = ctx
	_ = req
	return &emptypb.Empty{}, nil
}

func toProtoRoom(room *domain.Room) *pb.Room {
	if room == nil {
		return nil
	}
	return &pb.Room{
		Id:          room.ID.String(),
		Name:        room.Name,
		Description: room.Description,
		Type:        room.Type,
		OwnerId:     room.OwnerID.String(),
		MemberCount: int32(room.MemberCount), //nolint:gosec // safe: member count fits in int32
		CreatedAt:   timestamppb.New(room.CreatedAt),
		UpdatedAt:   timestamppb.New(room.UpdatedAt),
	}
}

func toProtoMessage(msg *domain.Message) *pb.Message {
	if msg == nil {
		return nil
	}
	protoMsg := &pb.Message{
		Id:             msg.ID.String(),
		RoomId:         msg.RoomID.String(),
		SenderId:       msg.SenderID.String(),
		SenderUsername: msg.SenderUsername,
		Content:        msg.Content,
		MessageType:    msg.MessageType,
		CreatedAt:      timestamppb.New(msg.CreatedAt),
	}
	if msg.ReplyTo != nil {
		protoMsg.ReplyTo = msg.ReplyTo.String()
	}
	return protoMsg
}
