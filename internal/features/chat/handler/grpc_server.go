package handler

import (
	"context"
	"io"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/zercle/zercle-go-template/api/pb"
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/features/chat/service"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	chatService *service.ChatService
}

func NewChatServer(chatService *service.ChatService) *ChatServer {
	return &ChatServer{chatService: chatService}
}

func (s *ChatServer) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Room, error) {
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return toProtoRoom(room), nil
}

func (s *ChatServer) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.Room, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	room, err := s.chatService.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return toProtoRoom(room), nil
}

func (s *ChatServer) UpdateRoom(ctx context.Context, req *pb.UpdateRoomRequest) (*pb.Room, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	room, err := s.chatService.UpdateRoom(ctx, roomID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return toProtoRoom(room), nil
}

func (s *ChatServer) DeleteRoom(ctx context.Context, req *pb.DeleteRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	if err := s.chatService.DeleteRoom(ctx, roomID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ChatServer) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	rooms, total, err := s.chatService.ListRooms(ctx, userID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	protoRooms := make([]*pb.Room, len(rooms))
	for i, room := range rooms {
		protoRooms[i] = toProtoRoom(room)
	}

	return &pb.ListRoomsResponse{
		Rooms: protoRooms,
		Total: int32(total),
	}, nil
}

func (s *ChatServer) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	if err := s.chatService.JoinRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ChatServer) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*emptypb.Empty, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	if err := s.chatService.LeaveRoom(ctx, roomID, userID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ChatServer) GetRoomMembers(ctx context.Context, req *pb.GetRoomMembersRequest) (*pb.GetRoomMembersResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	members, err := s.chatService.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, err
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

func (s *ChatServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
	}

	senderID, err := uuid.Parse(req.SenderId)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return toProtoMessage(message), nil
}

func (s *ChatServer) GetMessageHistory(ctx context.Context, req *pb.GetMessageHistoryRequest) (*pb.GetMessageHistoryResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, err
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
		return nil, err
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

func (s *ChatServer) ChatStream(stream pb.ChatService_ChatStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch payload := req.Payload.(type) {
		case *pb.ChatStreamRequest_Join:
			roomID, _ := uuid.Parse(payload.Join.RoomId)
			userID, _ := uuid.Parse(payload.Join.UserId)
			if err := s.chatService.JoinRoom(stream.Context(), roomID, userID); err != nil {
				return err
			}
		case *pb.ChatStreamRequest_Leave:
			roomID, _ := uuid.Parse(payload.Leave.RoomId)
			userID, _ := uuid.Parse(payload.Leave.UserId)
			if err := s.chatService.LeaveRoom(stream.Context(), roomID, userID); err != nil {
				return err
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
				return err
			}
			if err := stream.Send(&pb.ChatStreamResponse{
				Payload: &pb.ChatStreamResponse_Message{
					Message: toProtoMessage(sentMsg),
				},
			}); err != nil {
				return err
			}
		}
	}
}

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
		MemberCount: int32(room.MemberCount),
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
