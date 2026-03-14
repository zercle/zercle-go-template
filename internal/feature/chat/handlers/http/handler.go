package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/api/dtos/chat"
	"github.com/zercle/zercle-go-template/internal/feature/chat/domain"
	"github.com/zercle/zercle-go-template/internal/feature/chat/ports"
)

// Handler handles HTTP requests for chat.
type Handler struct {
	service ports.Service
}

// NewHandler creates a new HTTP chat handler.
func NewHandler(service ports.Service) *Handler {
	return &Handler{service: service}
}

// CreateRoom handles creating a new chat room.
func (h *Handler) CreateRoom(c *echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	ownerID, ok := userID.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	var req chat.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, idStr := range req.MemberIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		memberIDs = append(memberIDs, id)
	}

	input := ports.CreateRoomInput{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		OwnerID:     ownerID,
		MemberIDs:   memberIDs,
	}

	room, err := h.service.CreateRoom(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, toRoomResponse(room))
}

// ListRooms handles listing chat rooms.
func (h *Handler) ListRooms(c *echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	rooms, total, err := h.service.ListRooms(c.Request().Context(), uid, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	roomResponses := make([]*chat.RoomResponse, len(rooms))
	for i, r := range rooms {
		roomResponses[i] = toRoomResponse(r)
	}

	return c.JSON(http.StatusOK, chat.RoomListResponse{
		Rooms:  roomResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// GetRoom handles getting a single chat room.
func (h *Handler) GetRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	room, err := h.service.GetRoom(c.Request().Context(), roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "room not found")
	}

	return c.JSON(http.StatusOK, toRoomResponse(room))
}

// UpdateRoom handles updating a chat room.
func (h *Handler) UpdateRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	var req chat.UpdateRoomRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	room, err := h.service.UpdateRoom(c.Request().Context(), roomID, req.Name, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, toRoomResponse(room))
}

// DeleteRoom handles deleting a chat room.
func (h *Handler) DeleteRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	if err := h.service.DeleteRoom(c.Request().Context(), roomID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// JoinRoom handles joining a chat room.
func (h *Handler) JoinRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	userID := c.Get("user_id")
	if userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	if err := h.service.JoinRoom(c.Request().Context(), roomID, uid); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// LeaveRoom handles leaving a chat room.
func (h *Handler) LeaveRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	userID := c.Get("user_id")
	if userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	if err := h.service.LeaveRoom(c.Request().Context(), roomID, uid); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// GetRoomMembers handles getting room members.
func (h *Handler) GetRoomMembers(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	members, err := h.service.GetRoomMembers(c.Request().Context(), roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	memberResponses := make([]*chat.MemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = toMemberResponse(m)
	}

	return c.JSON(http.StatusOK, memberResponses)
}

// SendMessage handles sending a message.
func (h *Handler) SendMessage(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	userID := c.Get("user_id")
	if userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	var req chat.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	var replyTo *uuid.UUID
	if req.ReplyTo != nil {
		id, err := uuid.Parse(*req.ReplyTo)
		if err == nil {
			replyTo = &id
		}
	}

	input := ports.SendMessageInput{
		RoomID:      roomID,
		SenderID:    uid,
		Content:     req.Content,
		MessageType: req.MessageType,
		ReplyTo:     replyTo,
	}

	message, err := h.service.SendMessage(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, toMessageResponse(message))
}

// GetMessageHistory handles getting message history.
func (h *Handler) GetMessageHistory(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	var before *uuid.UUID
	beforeStr := c.QueryParam("before")
	if beforeStr != "" {
		id, err := uuid.Parse(beforeStr)
		if err == nil {
			before = &id
		}
	}

	messages, hasMore, err := h.service.GetMessageHistory(c.Request().Context(), roomID, limit, offset, before)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	messageResponses := make([]*chat.MessageResponse, len(messages))
	for i, m := range messages {
		messageResponses[i] = toMessageResponse(m)
	}

	return c.JSON(http.StatusOK, chat.MessageHistoryResponse{
		Messages: messageResponses,
		HasMore:  hasMore,
	})
}

func toRoomResponse(room *domain.Room) *chat.RoomResponse {
	if room == nil {
		return nil
	}
	return &chat.RoomResponse{
		ID:          room.ID.String(),
		Name:        room.Name,
		Description: room.Description,
		Type:        room.Type,
		OwnerID:     room.OwnerID.String(),
		MemberCount: room.MemberCount,
		CreatedAt:   room.CreatedAt,
		UpdatedAt:   room.UpdatedAt,
	}
}

func toMemberResponse(member *domain.RoomMember) *chat.MemberResponse {
	if member == nil {
		return nil
	}
	return &chat.MemberResponse{
		RoomID:      member.RoomID.String(),
		UserID:      member.UserID.String(),
		Username:    member.Username,
		DisplayName: member.DisplayName,
		AvatarURL:   member.AvatarURL,
		Role:        member.Role,
		JoinedAt:    member.JoinedAt,
	}
}

func toMessageResponse(message *domain.Message) *chat.MessageResponse {
	if message == nil {
		return nil
	}
	var replyTo *string
	if message.ReplyTo != nil {
		idStr := message.ReplyTo.String()
		replyTo = &idStr
	}
	return &chat.MessageResponse{
		ID:             message.ID.String(),
		RoomID:         message.RoomID.String(),
		SenderID:       message.SenderID.String(),
		SenderUsername: message.SenderUsername,
		Content:        message.Content,
		MessageType:    message.MessageType,
		ReplyTo:        replyTo,
		CreatedAt:      message.CreatedAt,
	}
}
