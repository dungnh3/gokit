package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/tikivn/ims-library/pkg/library/log"
	"go.uber.org/zap"
	"time"
)

type publisher struct {
	brokers        []string
	logger         *zap.Logger
	saramaProducer sarama.SyncProducer
	saramaConfig   *sarama.Config
}

type publisherOption func(producer *publisher) error

func defaultProducerConfig() (*publisher, error) {
	logger, err := log.DefaultConfig().Build()
	if err != nil {
		return nil, err
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true

	return &publisher{
		brokers:        []string{"127.0.0.1:9092"},
		logger:         logger,
		saramaProducer: nil,
		saramaConfig:   saramaConfig,
	}, nil
}

func NewPublisher(brokers []string, opts ...publisherOption) (*publisher, error) {
	p, err := defaultProducerConfig()
	if err != nil {
		return nil, err
	}

	opts = append(opts, withBroker(brokers))
	for _, opt := range opts {
		_ = opt(p)
	}
	p.saramaProducer, err = sarama.NewSyncProducer(p.brokers, p.saramaConfig)
	return p, nil
}

func withBroker(broker []string) publisherOption {
	return func(p *publisher) error {
		p.brokers = broker
		return nil
	}
}

func WithProducerConfig(config *sarama.Config) publisherOption {
	return func(p *publisher) error {
		p.saramaConfig = config
		return nil
	}
}

func (p *publisher) Publish(ctx context.Context, topic string, message []byte, opts ...publisherMessageOption) error {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       nil,
		Value:     nil,
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Time{},
	}

	opts = append(opts, withPublisherMsgValue(message))
	for _, opt := range opts {
		if err := opt(msg); err != nil {
			return err
		}
	}

	partition, offset, err := p.saramaProducer.SendMessage(msg)
	if err != nil {
		p.logger.Error("Publish message to topic failed",
			zap.Any("with error", err),
		)
		return err
	}
	p.logger.Info("publish messages success with it's information",
		zap.Any("topic", topic),
		zap.Any("partition", partition),
		zap.Any("offset", offset),
	)
	return nil
}

type publisherMessageOption func(message *sarama.ProducerMessage) error

func withPublisherMsgValue(message []byte) publisherMessageOption {
	return func(msg *sarama.ProducerMessage) error {
		msg.Value = sarama.ByteEncoder(message)
		return nil
	}
}

func WithPublisherMsgKey(key []byte) publisherMessageOption {
	return func(msg *sarama.ProducerMessage) error {
		msg.Key = sarama.ByteEncoder(key)
		return nil
	}
}

func WithPublisherMsgPartition(partition int32) publisherMessageOption {
	return func(msg *sarama.ProducerMessage) error {
		msg.Partition = partition
		return nil
	}
}

func WithPublisherMsgTimestamp(timestamp time.Time) publisherMessageOption {
	return func(msg *sarama.ProducerMessage) error {
		msg.Timestamp = timestamp
		return nil
	}
}

func WithPublisherMsgHeader(headers map[string]string) publisherMessageOption {
	return func(msg *sarama.ProducerMessage) error {
		recordHeaders := make([]sarama.RecordHeader, 0, len(headers))
		for k, v := range headers {
			rh := sarama.RecordHeader{
				Key:   sarama.ByteEncoder(k),
				Value: sarama.ByteEncoder(v),
			}
			recordHeaders = append(recordHeaders, rh)
		}
		msg.Headers = recordHeaders
		return nil
	}
}
