package sse

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
)

// Handler manages Server-Sent Events for real-time chat updates.
type Handler struct {
	valkeyClient valkey.PubSubClient
}

// NewHandler creates a new SSE handler with the given Valkey client.
func NewHandler(valkeyClient valkey.PubSubClient) *Handler {
	return &Handler{valkeyClient: valkeyClient}
}

// Event represents a Server-Sent Event with type and payload.
type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// HandleSSE handles SSE connections for a chat room.
func (h *Handler) HandleSSE(c *echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "room ID required")
	}

	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("failed to parse room ID: %w", err)
	}

	_ = roomUUID

	userID, err := h.getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	channel := fmt.Sprintf("room:%s", roomID)
	pubsub, err := h.valkeyClient.Subscribe(c.Request().Context(), channel)
	if err != nil {
		return fmt.Errorf("failed to subscribe to room: %w", err)
	}
	defer func() { _ = pubsub.Close() }()

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	if err := h.sendEvent(c, "connected", map[string]string{
		"room_id": roomID,
		"user_id": userID.String(),
	}); err != nil {
		return fmt.Errorf("failed to send connected event: %w", err)
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
			if err := sonic.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}
			if err := h.sendEvent(c, event.Type, event.Payload); err != nil {
				return nil
			}
		}
	}
}

// PublishMessage publishes a message event to a room's SSE channel.
func (h *Handler) PublishMessage(ctx context.Context, roomID string, message any) error {
	event := Event{
		Type:    "message",
		Payload: message,
	}
	data, err := sonic.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	if err := h.valkeyClient.Publish(ctx, fmt.Sprintf("room:%s", roomID), string(data)); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (h *Handler) sendEvent(c *echo.Context, eventType string, data any) error {
	event := Event{
		Type:    eventType,
		Payload: data,
	}
	jsonData, err := sonic.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	if _, err := fmt.Fprintf(c.Response(), "data: %s\n\n", jsonData); err != nil {
		return fmt.Errorf("failed to send SSE event: %w", err)
	}
	return nil
}

func (h *Handler) getUserID(c *echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	return userID, nil
}
