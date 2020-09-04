package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
)

var (
	// RequiredAcks will be used in Kafka configs
	// to set the 'RequiredAcks' value.
	RequireAcks = sarama.WaitForAll

	// ErrTopicNameIsRequired ..
	ErrBrokerHostIsRequired = errors.New("broker host is required")
	// ErrTopicNameIsRequired ..
	ErrTopicNameIsRequired = errors.New("topic name is required")
	// ErrChannelMsgIsClosed ..
	ErrChannelMsgIsClosed = errors.New("channel message is close")
)

// ProducerConfig is a configuration for kafka producer
type ProducerConfig struct {
	Brokers      []string
	Topic        string
	MaxRetry     int
	Partition    int32
	saramaConfig *sarama.Config
}

// ConsumerConfig is a configuration for kafka consumer
type ConsumerConfig struct {
	Brokers      []string
	BrokersRetry []string
	BrokersDLQ   []string
	Topic        string
	MaxRetry     int
	saramaConfig *sarama.Config
}
