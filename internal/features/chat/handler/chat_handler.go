package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/features/chat/dto"
	"github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/middleware"
)

type ChatHandler struct {
	chatService service.ChatServiceInterface
}

func NewChatHandler(chatService service.ChatServiceInterface) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// CreateRoom godoc
// @Summary Create a new chat room
// @Description Create a new chat room with specified members
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateRoomRequest true "Room details"
// @Success 201 {object} dto.RoomResponse "Room created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /chat/rooms [post]
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

// GetRoom godoc
// @Summary Get room details
// @Description Get details of a specific chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Success 200 {object} dto.RoomResponse "Room details"
// @Failure 400 {object} map[string]string "Invalid room ID"
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/{id} [get]
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

// ListRooms godoc
// @Summary List user's chat rooms
// @Description Get all chat rooms the authenticated user is a member of
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param limit query int false "Number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} dto.ListRoomsResponse "List of rooms"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /chat/rooms [get]
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

// UpdateRoom godoc
// @Summary Update room details
// @Description Update name and description of a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Param request body dto.UpdateRoomRequest true "Updated room details"
// @Success 200 {object} dto.RoomResponse "Room updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /chat/rooms/{id} [put]
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

// DeleteRoom godoc
// @Summary Delete a chat room
// @Description Delete a chat room (owner only)
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Success 204 "Room deleted successfully"
// @Failure 400 {object} map[string]string "Invalid room ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /chat/rooms/{id} [delete]
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

// JoinRoom godoc
// @Summary Join a chat room
// @Description Add the authenticated user to a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Success 204 "Successfully joined room"
// @Failure 400 {object} map[string]string "Invalid room ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /chat/rooms/{id}/join [post]
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

// LeaveRoom godoc
// @Summary Leave a chat room
// @Description Remove the authenticated user from a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Success 204 "Successfully left room"
// @Failure 400 {object} map[string]string "Invalid room ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /chat/rooms/{id}/leave [post]
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

// GetRoomMembers godoc
// @Summary Get room members
// @Description Get all members of a specific chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Success 200 {array} dto.RoomResponse "List of room members"
// @Failure 400 {object} map[string]string "Invalid room ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /chat/rooms/{id}/members [get]
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

// SendMessage godoc
// @Summary Send a message to a room
// @Description Send a new message to a specific chat room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Param request body dto.SendMessageRequest true "Message content"
// @Success 201 {object} dto.MessageResponse "Message sent successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /chat/rooms/{id}/messages [post]
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

// GetMessageHistory godoc
// @Summary Get message history
// @Description Get paginated message history for a room
// @Tags chat
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Room ID" format(uuid)
// @Param limit query int false "Number of messages" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Param before query string false "Get messages before this message ID" format(uuid)
// @Success 200 {object} dto.GetMessagesResponse "Message history"
// @Failure 400 {object} map[string]string "Invalid room ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /chat/rooms/{id}/messages [get]
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
