package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Event represents an event envelope emitted to Kafka topics.
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// KafkaProducer encapsulates writing events to Kafka topics with connection pooling and batching.
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer initializes a kafka writer with snappy/gzip compression and retries.
func NewKafkaProducer(brokers []string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		Async:        false,
		RequiredAcks: kafka.RequireOne,
	}
	return &KafkaProducer{writer: writer}
}

// Publish publishes a message with a specific routing key to guarantee partition-level ordering.
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka payload: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Time:  time.Now().UTC(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write kafka message to topic %s: %w", topic, err)
	}
	return nil
}

// Close flushes buffered messages and closes the writer.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// EventHandler defines the signature for Kafka event handlers.
type EventHandler func(ctx context.Context, key []byte, value []byte) error

// KafkaConsumer reads from a Kafka topic within a consumer group with graceful cancellation.
type KafkaConsumer struct {
	reader  *kafka.Reader
	handler EventHandler
}

// NewKafkaConsumer creates a new KafkaConsumer for a specific topic and groupID.
func NewKafkaConsumer(brokers []string, topic string, groupID string, handler EventHandler) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
	})
	return &KafkaConsumer{
		reader:  reader,
		handler: handler,
	}
}

// Start listens continuously for messages until context is cancelled.
func (c *KafkaConsumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Kafka Consumer] Context cancelled, stopping reader")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Kafka Consumer] Error fetching message: %v", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if err := c.handler(ctx, msg.Key, msg.Value); err != nil {
				log.Printf("[Kafka Consumer] Handler error on topic %s partition %d offset %d: %v",
					msg.Topic, msg.Partition, msg.Offset, err)
				// Retry or dead-letter strategy here
			} else {
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("[Kafka Consumer] Failed to commit message: %v", err)
				}
			}
		}
	}
}

// Close terminates the reader.
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
