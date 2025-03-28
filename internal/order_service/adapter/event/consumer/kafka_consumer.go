package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
)

// KafkaConsumer is an implementation of the domain's EventConsumer interface
type KafkaConsumer struct {
	brokers []string
	groupID string
	readers map[string]*kafka.Reader
}

// NewKafkaConsumer creates a new Kafka event consumer
func NewKafkaConsumer(brokers []string, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		brokers: brokers,
		groupID: groupID,
		readers: make(map[string]*kafka.Reader),
	}
}

// Subscribe subscribes to events from a topic
func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler func(event service.Event) error) error {
	// Create reader if it doesn't exist for this topic
	if _, exists := c.readers[topic]; !exists {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        c.brokers,
			Topic:          topic,
			GroupID:        c.groupID,
			MinBytes:       10e3,        // 10KB
			MaxBytes:       10e6,        // 10MB
			MaxWait:        1 * time.Second,
			CommitInterval: 1 * time.Second,
			StartOffset:    kafka.FirstOffset,
		})
		
		c.readers[topic] = reader
	}
	
	// Start consuming in a goroutine
	go func() {
		reader := c.readers[topic]
		
		// Create a derived context that can be cancelled when we need to stop
		consumerCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		
		// Listen for context cancellation to close the reader
		go func() {
			<-consumerCtx.Done()
			if err := reader.Close(); err != nil {
				log.Printf("Error closing Kafka reader: %v", err)
			}
		}()
		
		// Consume messages
		for {
			message, err := reader.FetchMessage(consumerCtx)
			if err != nil {
				// Context cancelled or other error
				if consumerCtx.Err() != nil {
					return // Exit gracefully when context is cancelled
				}
				
				log.Printf("Error fetching message from Kafka: %v", err)
				time.Sleep(1 * time.Second) // Back off on error
				continue
			}
			
			// Parse the event
			var event service.Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Error unmarshalling Kafka message: %v", err)
				// Commit the message anyway to avoid getting stuck on a bad message
				if err := reader.CommitMessages(consumerCtx, message); err != nil {
					log.Printf("Error committing message: %v", err)
				}
				continue
			}
			
			// Handle the event
			if err := handler(event); err != nil {
				log.Printf("Error handling event: %v", err)
				// Depending on business logic, we might not commit the message here
				// to allow for retry
				continue
			}
			
			// Commit the message
			if err := reader.CommitMessages(consumerCtx, message); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}()
	
	return nil
}

// Close closes all Kafka readers
func (c *KafkaConsumer) Close() error {
	var lastErr error
	
	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			lastErr = fmt.Errorf("error closing reader for topic %s: %w", topic, err)
			log.Printf("%v", lastErr)
		}
	}
	
	return lastErr
}
