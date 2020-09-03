package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

// producer is an experimental Publisher that provides an implementation for
// Kafka using the Shopify/sarama library.
type producer struct {
	saramaProducer sarama.SyncProducer
	topic          string
}

// NewPublisher will initiate a new experimental Kafka producer.
func NewPublisher(cfg *ProducerConfig) (*producer, error) {
	var err error

	if len(cfg.Brokers) == 0 {
		return nil, ErrBrokerHostIsRequired
	}

	if len(cfg.Topic) == 0 {
		return nil, ErrTopicNameIsRequired
	}
	p := new(producer)
	p.topic = cfg.Topic

	sconfig := cfg.saramaConfig
	if sconfig == nil {
		sconfig = sarama.NewConfig()
		sconfig.Version = sarama.V2_4_0_0
		sconfig.Producer.Retry.Max = cfg.MaxRetry
		sconfig.Producer.RequiredAcks = sarama.WaitForAll
	}
	sconfig.Producer.Return.Successes = true
	if sconfig.Producer.Retry.Max == 0 {
		sconfig.Producer.Retry.Max = 5
	}
	p.saramaProducer, err = sarama.NewSyncProducer(cfg.Brokers, sconfig)
	return p, err
}

// PublishRaw ..
func (p *producer) Publish(ctx context.Context, key []byte, msg []byte) error {
	message := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.ByteEncoder(key),
		Value:     sarama.ByteEncoder(msg),
		Timestamp: time.Time{},
	}
	partition, offset, err := p.saramaProducer.SendMessage(message)
	log.Printf("Message is stored in topic(%v)/partition(%v)/offset(%v) \n", p.topic, partition, offset)
	return err
}

// Close ..
func (p *producer) Close() error {
	return p.saramaProducer.Close()
}
