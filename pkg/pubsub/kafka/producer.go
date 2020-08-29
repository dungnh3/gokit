package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/dungnh3/gokit/pkg/log/level"
)

type Publisher struct {
	producer sarama.SyncProducer
	topic    string
}

func NewPublisher(cfg *Config) (*Publisher, error) {
	if len(cfg.Topic) == 0 {
		return nil, ErrTopicIsRequired
	}
	p := new(Publisher)
	p.topic = cfg.Topic[0]

	if cfg.Config == nil {
		cfg.Config = sarama.NewConfig()
	}
	cfg.Config.Producer.Retry.Max = cfg.MaxRetry
	cfg.Config.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Config.Producer.Return.Successes = true
	var err error
	p.producer, err = sarama.NewSyncProducer(cfg.BrokerHost.Cluster, cfg.Config)
	return p, err
}

func (p *Publisher) Publish(ctx context.Context, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		level.Error(ctx).F("publish message into %v topic failed, error %v", err)
		return err
	}
	level.Info(ctx).F("publish message into %v topic, %v partition, %v offset success", p.topic, partition, offset)
	return nil
}
