package kafka

import "github.com/Shopify/sarama"

type MetaData struct {
	MaxRetry     int `json:"max_retry"`
	CounterRetry int `json:"counter_retry"`
}

type ConsumerMessage struct {
	MetaData MetaData                `json:"meta_data"`
	Msg      *sarama.ConsumerMessage `json:"msg"`
}
