package dbs

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/quangdangfit/goticket/config"
)

// KafkaPublisher publishes messages to a topic.
type KafkaPublisher interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

type kafkaWriter struct {
	w       *kafka.Writer
	brokers []string
}

// NewKafkaPublisher returns a multi-topic publisher backed by kafka-go.
func NewKafkaPublisher(cfg config.KafkaConfig) KafkaPublisher {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           0,
	}
	return &kafkaWriter{w: w, brokers: cfg.Brokers}
}

func (k *kafkaWriter) Publish(ctx context.Context, topic string, key, value []byte) error {
	return k.w.WriteMessages(ctx, kafka.Message{Topic: topic, Key: key, Value: value})
}

func (k *kafkaWriter) Close() error { return k.w.Close() }
