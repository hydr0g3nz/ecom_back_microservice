package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
)

// KafkaProducer is an implementation of the domain's EventProducer interface
type KafkaProducer struct {
	writer *kafka.Writer
	addr   string
}

// NewKafkaProducer creates a new Kafka event producer
func NewKafkaProducer(brokers []string) *KafkaProducer {
	addr := brokers[0] // Use the first broker for connection
	
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(addr),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireOne,
		Compression:            kafka.Snappy,
		ReadTimeout:            10 * time.Second,
		WriteTimeout:           10 * time.Second,
	}
	
	return &KafkaProducer{
		writer: writer,
		addr:   addr,
	}
}

// Publish publishes an event to the Kafka message broker
func (p *KafkaProducer) Publish(ctx context.Context, topic string, event service.Event) error {
	// Marshal event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Create message
	message := kafka.Message{
		Topic: topic,
		Key:   []byte(string(event.Type)), // Use event type as the key
		Value: payload,
		Time:  time.Now(),
	}
	
	// Set the topic in the writer
	p.writer.Topic = topic
	
	// Write message
	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	
	return nil
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
