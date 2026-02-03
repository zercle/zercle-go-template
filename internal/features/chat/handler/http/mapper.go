package http

import (
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/features/chat/dto"
)

func ToRoomResponse(room *domain.Room) *dto.RoomResponse {
	return &dto.RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Type:        room.Type,
		OwnerID:     room.OwnerID,
		MemberCount: room.MemberCount,
		CreatedAt:   room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToMessageResponse(msg *domain.Message) *dto.MessageResponse {
	resp := &dto.MessageResponse{
		ID:             msg.ID,
		RoomID:         msg.RoomID,
		SenderID:       msg.SenderID,
		SenderUsername: msg.SenderUsername,
		Content:        msg.Content,
		MessageType:    msg.MessageType,
		CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if msg.ReplyTo != nil {
		resp.ReplyTo = msg.ReplyTo.String()
	}
	return resp
}
