package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"time"
)

type Producer interface {
	Publish(ctx context.Context, key, msg []byte) error
}

type producer struct {
	brokers []string
	srmProducer sarama.SyncProducer
}

func NewProducer(brokers []string) (*producer, error) {
	return nil, nil
}

func (p *producer) Publish(ctx context.Context, topic string, key []byte, msg []byte) error {
	p.srmProducer.SendMessage(sarama.ProducerMessage{
		Topic:     topic,
		Key:       nil,
		Value:     nil,
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Time{},
	}) 
}
