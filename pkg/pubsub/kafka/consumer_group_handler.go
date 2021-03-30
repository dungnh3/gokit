package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type consumerGroupHandler struct {
	ctx    context.Context
	logger *zap.Logger

	readyChan   chan bool
	messageChan chan *consumerGroupMessage

	maxRetry int
}

// Setup takes no action (required for ConsumerGroupHandler interface).
func (cgh *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as readyChan
	cgh.logger.Info("close ready channel!")
	close(cgh.readyChan)
	return nil
}

// Cleanup takes no action
func (cgh *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		select {
		case <-cgh.ctx.Done():
			cgh.logger.Info("Receive done signal from context. Stop claim message from topic..")
			return nil
		case cgh.messageChan <- &consumerGroupMessage{
			message:      msg,
			session:      session,
			maxRetry:     cgh.maxRetry,
			counterRetry: 0}:
		}
	}
	return nil
}
