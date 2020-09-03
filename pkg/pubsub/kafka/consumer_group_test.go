package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
	"log"
	"testing"
)

var consumerConfig *ConsumerConfig

func init() {
	// ctx, cancel := context.WithCancel(context.Background())
	consumerConfig = &ConsumerConfig{
		Brokers:      []string{"127.0.0.1:9092"},
		BrokersRetry: []string{"127.0.0.1:9092"},
		BrokersDLQ:   []string{"127.0.0.1:9092"},
		Topic:        "sample_topic",
		MaxRetry:     3,
		Config:       sarama.NewConfig(),
	}
}

func TestNewKafkaConsumerGroup(t *testing.T) {
	consumerConfig.Config.Version = sarama.V2_4_0_0
	consumerConfig.Config.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumerGroups, err := NewKafkaConsumerGroup(consumerConfig, "sample_group", processDLQ)
	if err != nil {
		return
	}
	consumerGroups.Start()
}

func process(msg *sarama.ConsumerMessage) error {
	log.Printf("msg => %v", string(msg.Value))
	return nil
}

func processDLQ(msg *sarama.ConsumerMessage) error {
	log.Printf("process msg => %v", string(msg.Value))
	return errors.New("fake error")
}
