package kafka

import (
	"context"
	"encoding/json"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/publisher"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer handles publishing events to Kafka
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new instance of KafkaProducer
func NewKafkaProducer(brokers []string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// PublishMessage publishes a message to Kafka
func (p *KafkaProducer) PublishMessage(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
}

// Close closes the Kafka writer
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// CreateOrderEventPublisher creates an OrderEventPublisher using Kafka
func (p *KafkaProducer) CreateOrderEventPublisher() publisher.OrderEventPublisher {
	return publisher.NewKafkaOrderEventPublisher([]string{p.writer.Addr.String()})
}
