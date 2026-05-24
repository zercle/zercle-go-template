package di

import (
	"github.com/samber/do/v2"

	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/features/chat/handler/http"
	"github.com/zercle/zercle-go-template/internal/features/chat/handler/sse"
	"github.com/zercle/zercle-go-template/internal/features/chat/messaging"
	"github.com/zercle/zercle-go-template/internal/features/chat/service"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db/postgres"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// ProvideRoomRepository creates and provides a room repository.
func ProvideRoomRepository(i do.Injector) (domain.RoomRepository, error) {
	db := do.MustInvoke[*postgres.DB](i)
	return postgres.NewRoomRepository(db), nil
}

// ProvideMessageRepository creates and provides a message repository.
func ProvideMessageRepository(i do.Injector) (domain.MessageRepository, error) {
	db := do.MustInvoke[*postgres.DB](i)
	return postgres.NewMessageRepository(db), nil
}

// ProvidePubSubService creates and provides the PubSub messaging service.
func ProvidePubSubService(i do.Injector) (messaging.PubSubServiceInterface, error) {
	client := do.MustInvoke[*valkey.Client](i)
	logger := do.MustInvoke[*telemetry.Logger](i)
	return messaging.New(client, &logger.Logger), nil
}

// ProvideChatService creates and provides the chat service.
func ProvideChatService(i do.Injector) (service.ChatServiceInterface, error) {
	roomRepo := do.MustInvoke[domain.RoomRepository](i)
	messageRepo := do.MustInvoke[domain.MessageRepository](i)
	pubsub := do.MustInvoke[messaging.PubSubServiceInterface](i)
	logger := do.MustInvoke[*telemetry.Logger](i)

	return service.NewChatServiceWithPubSub(roomRepo, messageRepo, pubsub, &logger.Logger), nil
}

// ProvideChatHandler creates and provides the chat HTTP handler.
func ProvideChatHandler(i do.Injector) (*http.ChatHandler, error) {
	chatSvc := do.MustInvoke[service.ChatServiceInterface](i)
	return http.NewChatHandler(chatSvc), nil
}

// ProvideSSEHandler creates and provides the SSE handler.
func ProvideSSEHandler(i do.Injector) (*sse.Handler, error) {
	client := do.MustInvoke[*valkey.Client](i)
	return sse.NewHandler(client), nil
}

// RegisterChatProviders registers all chat-related DI providers.
func RegisterChatProviders(i do.Injector) {
	do.Provide(i, ProvideRoomRepository)
	do.Provide(i, ProvideMessageRepository)
	do.Provide(i, ProvidePubSubService)
	do.Provide(i, ProvideChatService)
	do.Provide(i, ProvideChatHandler)
	do.Provide(i, ProvideSSEHandler)
}
