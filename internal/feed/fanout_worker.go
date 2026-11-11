package feed

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofrs/uuid/v5"

	"github.com/Rizwank123/chatApp/internal/broker"
	"github.com/Rizwank123/chatApp/internal/domain"
)

// PostCreatedPayload represents the Kafka event payload for a new post.
type PostCreatedPayload struct {
	PostID      uuid.UUID `json:"post_id"`
	PersonnelID uuid.UUID `json:"personnel_id"`
	Caption     string    `json:"caption"`
	MediaURL    []string  `json:"media_url"`
	Visibility  string    `json:"visibility"`
}

// FanoutWorker consumes PostCreated events from Kafka and drives fanout-on-write into Redis timelines.
type FanoutWorker struct {
	consumer   *broker.KafkaConsumer
	feedEngine *HybridFeedEngine
	postRepo   domain.PostRepository
}

// NewFanoutWorker sets up the Kafka consumer for the post fanout group.
func NewFanoutWorker(
	brokers []string,
	feedEngine *HybridFeedEngine,
	postRepo domain.PostRepository,
) *FanoutWorker {
	w := &FanoutWorker{
		feedEngine: feedEngine,
		postRepo:   postRepo,
	}

	w.consumer = broker.NewKafkaConsumer(
		brokers,
		"social.posts.created",
		"feed-fanout-workers",
		w.handlePostCreated,
	)

	return w
}

// Start begins processing events asynchronously.
func (w *FanoutWorker) Start(ctx context.Context) {
	log.Printf("[FanoutWorker] Starting Kafka consumer for social.posts.created...")
	go w.consumer.Start(ctx)
}

// Close gracefully stops the consumer.
func (w *FanoutWorker) Close() error {
	return w.consumer.Close()
}

func (w *FanoutWorker) handlePostCreated(ctx context.Context, key []byte, value []byte) error {
	var payload PostCreatedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		log.Printf("[FanoutWorker] Failed to unmarshal PostCreatedPayload: %v", err)
		return err
	}

	// Fetch full post
	post, err := w.postRepo.FindPostByID(ctx, payload.PostID)
	if err != nil {
		log.Printf("[FanoutWorker] Failed to find post %s: %v", payload.PostID, err)
		return err
	}

	// Execute fanout logic
	if err := w.feedEngine.FanoutPost(ctx, &post); err != nil {
		log.Printf("[FanoutWorker] Error during fanout for post %s: %v", post.ID, err)
		return err
	}

	log.Printf("[FanoutWorker] Successfully completed fanout for post %s by author %s", post.ID, post.PersonnelID)
	return nil
}
