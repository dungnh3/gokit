package kafka

import "github.com/Shopify/sarama"

type consumerGroupMessage struct {
	message *sarama.ConsumerMessage
	session sarama.ConsumerGroupSession

	maxRetry     int
	counterRetry int
}

func (cgm *consumerGroupMessage) Message() []byte {
	return cgm.message.Value
}

func (cgm *consumerGroupMessage) Key() []byte {
	return cgm.message.Key
}

func (cgm *consumerGroupMessage) Done() error {
	cgm.session.MarkMessage(cgm.message, "")
	return nil
}
