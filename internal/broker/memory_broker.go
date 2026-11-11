package broker

import (
	"context"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type inMemoryBroker struct {
	messages chan domain.MessageEvent
}

func NewInMemoryBroker() domain.MessageBroker {
	return &inMemoryBroker{
		messages: make(chan domain.MessageEvent, 256),
	}
}

func (b *inMemoryBroker) PublishMessage(ctx context.Context, event domain.MessageEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case b.messages <- event:
		return nil
	}
}

func (b *inMemoryBroker) SubscribeMessages(ctx context.Context) (<-chan domain.MessageEvent, error) {
	out := make(chan domain.MessageEvent)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-b.messages:
				select {
				case <-ctx.Done():
					return
				case out <- event:
				}
			}
		}
	}()
	return out, nil
}
