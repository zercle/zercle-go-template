package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/chat/dto"
	"github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/shared/middleware"
)

type ChatHandler struct {
	chatService service.ChatServiceInterface
}

func NewChatHandler(chatService service.ChatServiceInterface) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) CreateRoom(c *echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var req dto.CreateRoomRequest
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

	input := service.CreateRoomInput{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		OwnerID:     userID,
		MemberIDs:   memberIDs,
	}

	room, err := h.chatService.CreateRoom(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, ToRoomResponse(room))
}

func (h *ChatHandler) GetRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	room, err := h.chatService.GetRoom(c.Request().Context(), roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "room not found")
	}

	return c.JSON(http.StatusOK, ToRoomResponse(room))
}

func (h *ChatHandler) ListRooms(c *echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	limit := 20
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed := parseIntDefault(l, 20); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}

	rooms, total, err := h.chatService.ListRooms(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	roomResponses := make([]*dto.RoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = ToRoomResponse(room)
	}

	return c.JSON(http.StatusOK, dto.ListRoomsResponse{
		Rooms: roomResponses,
		Total: total,
	})
}

func (h *ChatHandler) UpdateRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	var req dto.UpdateRoomRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	room, err := h.chatService.UpdateRoom(c.Request().Context(), roomID, req.Name, req.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, ToRoomResponse(room))
}

func (h *ChatHandler) DeleteRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	if err := h.chatService.DeleteRoom(c.Request().Context(), roomID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ChatHandler) JoinRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.chatService.JoinRoom(c.Request().Context(), roomID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ChatHandler) LeaveRoom(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if err := h.chatService.LeaveRoom(c.Request().Context(), roomID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ChatHandler) GetRoomMembers(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	members, err := h.chatService.GetRoomMembers(c.Request().Context(), roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, members)
}

func (h *ChatHandler) SendMessage(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var req dto.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
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
		SenderID:    userID,
		Content:     req.Content,
		MessageType: req.MessageType,
		ReplyTo:     replyTo,
	}

	message, err := h.chatService.SendMessage(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, ToMessageResponse(message))
}

func (h *ChatHandler) GetMessageHistory(c *echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	limit := 50
	offset := 0
	var before *uuid.UUID

	if l := c.QueryParam("limit"); l != "" {
		if parsed := parseIntDefault(l, 50); parsed > 0 {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		offset = parseIntDefault(o, 0)
	}
	if b := c.QueryParam("before"); b != "" {
		if id, err := uuid.Parse(b); err == nil {
			before = &id
		}
	}

	messages, hasMore, err := h.chatService.GetMessageHistory(c.Request().Context(), roomID, limit, offset, before)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	msgResponses := make([]*dto.MessageResponse, len(messages))
	for i, msg := range messages {
		msgResponses[i] = ToMessageResponse(msg)
	}

	return c.JSON(http.StatusOK, dto.GetMessagesResponse{
		Messages: msgResponses,
		HasMore:  hasMore,
	})
}

func parseIntDefault(s string, defaultVal int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
