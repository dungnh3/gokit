package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
)

const (
	RetryTopic    = "%v_%v"
	ConsumerGroup = "%v_%v"
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

type ProducerConfig struct {
	Brokers   []string
	Topic     string
	MaxRetry  int
	Partition int32
	Config    *sarama.Config
}

type ConsumerConfig struct {
	Brokers      []string
	BrokersRetry []string
	BrokersDLQ   []string
	Topic        string
	MaxRetry     int
	Config       *sarama.Config
}
