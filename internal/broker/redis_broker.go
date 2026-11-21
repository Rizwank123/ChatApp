package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/Rizwank123/chatApp/internal/domain"
)

const chatMessageChannel = "chat:message-events"

// RedisMessageBroker distributes realtime delivery events to API replicas.
// PostgreSQL remains the durable source of truth for every message.
type RedisMessageBroker struct {
	client *redis.Client
}

func NewRedisMessageBroker(client *redis.Client) domain.MessageBroker {
	return &RedisMessageBroker{client: client}
}

func (b *RedisMessageBroker) PublishMessage(ctx context.Context, event domain.MessageEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode message event: %w", err)
	}
	return b.client.Publish(ctx, chatMessageChannel, payload).Err()
}

func (b *RedisMessageBroker) SubscribeMessages(ctx context.Context) (<-chan domain.MessageEvent, error) {
	pubsub := b.client.Subscribe(ctx, chatMessageChannel)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("subscribe to message events: %w", err)
	}

	output := make(chan domain.MessageEvent, 64)
	go func() {
		defer close(output)
		defer pubsub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-pubsub.Channel():
				if !ok {
					return
				}
				var event domain.MessageEvent
				if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
					continue
				}
				select {
				case output <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return output, nil
}
