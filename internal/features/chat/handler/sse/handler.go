package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
)

type Handler struct {
	valkeyClient valkey.PubSubClient
}

func NewHandler(valkeyClient valkey.PubSubClient) *Handler {
	return &Handler{valkeyClient: valkeyClient}
}

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (h *Handler) HandleSSE(c *echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "room ID required")
	}

	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room ID")
	}

	_ = roomUUID

	userID, err := h.getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	channel := fmt.Sprintf("room:%s", roomID)
	pubsub := h.valkeyClient.Subscribe(c.Request().Context(), channel)
	defer func() { _ = pubsub.Close() }()

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	if err := h.sendEvent(c, "connected", map[string]string{
		"room_id": roomID,
		"user_id": userID.String(),
	}); err != nil {
		return err
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case <-ticker.C:
			if err := h.sendEvent(c, "ping", map[string]string{}); err != nil {
				return nil
			}
		case msg, ok := <-pubsub.Channel():
			if !ok {
				return nil
			}
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}
			if err := h.sendEvent(c, event.Type, event.Payload); err != nil {
				return nil
			}
		}
	}
}

func (h *Handler) PublishMessage(ctx context.Context, roomID string, message interface{}) error {
	event := Event{
		Type:    "message",
		Payload: message,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return h.valkeyClient.Publish(ctx, fmt.Sprintf("room:%s", roomID), string(data))
}

func (h *Handler) sendEvent(c *echo.Context, eventType string, data interface{}) error {
	event := Event{
		Type:    eventType,
		Payload: data,
	}
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.Response(), "data: %s\n\n", jsonData)
	return err
}

func (h *Handler) getUserID(c *echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	return userID, nil
}
