package kafka

import "github.com/Shopify/sarama"

type MetaData struct {
	MaxRetry     int `json:"max_retry"`
	CounterRetry int `json:"counter_retry"`
}

type consumerMessage struct {
	MetaData              MetaData                `json:"meta_data"`
	saramaConsumerMessage *sarama.ConsumerMessage `json:"msg"`
}
