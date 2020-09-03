package kafka

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type consumerGroupHandler struct {
	ctx            context.Context
	maxRetry       int
	retryPublisher *producer
	dlqPublisher   *producer
	handler        func(msg *sarama.ConsumerMessage) error
}

func (cgh *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (cgh *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		//session.MarkMessage(msg, "")
		//continue
		err := cgh.handler(msg)
		if err != nil {
			go func() {
				// TODO: push msg (metadata + raw msg) into retry topic
				log.Printf("handler msg from topic failed, error: %v, consider push msg into retry topic %v \n", err, cgh.retryPublisher.topic)
				consumerMsg := &ConsumerMessage{
					MetaData: MetaData{
						MaxRetry:     cgh.maxRetry,
						CounterRetry: 0,
					},
					Msg: msg,
				}
				newValue, _ := json.Marshal(consumerMsg)

				err = cgh.retryPublisher.Publish(cgh.ctx, nil, newValue)
				if err != nil {
					// TODO: push raw msg into dlq topic
					log.Printf("push msg to retry topic failed, error: %v, consider push msg into dlq topic %v \n", err, cgh.dlqPublisher.topic)
					err = cgh.dlqPublisher.Publish(cgh.ctx, msg.Key, msg.Value)
					if err != nil {
						log.Printf("push msg to dlq topic failed, error: %v \n", err)
					}
				}
			}()
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

type consumerGroupRetryHandler struct {
	*consumerGroupHandler
}

func (cgrh *consumerGroupRetryHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		//session.MarkMessage(msg, "")
		//continue
		// TODO: parse msg(metadata + raw msg) to get raw msg
		var err error
		var consumerMsg ConsumerMessage
		err = json.Unmarshal(msg.Value, &consumerMsg)
		log.Printf("msg value => %v \n", string(consumerMsg.Msg.Value))
		if err != nil {
			log.Printf("unmarshal consumer msg failed, error %v", err)
			continue
		}

		err = cgrh.consumerGroupHandler.handler(consumerMsg.Msg)
		if err != nil {
			go func() {
				maxRetry := consumerMsg.MetaData.MaxRetry
				counterRetry := consumerMsg.MetaData.CounterRetry

				if counterRetry >= maxRetry {
					// TODO: threshold retry, consider push msg to dlq
					err = cgrh.consumerGroupHandler.dlqPublisher.Publish(cgrh.consumerGroupHandler.ctx, consumerMsg.Msg.Key, consumerMsg.Msg.Value)
					if err != nil {
						log.Printf("push msg to dlq topic failed, error: %v \n", err)
					}
					log.Println("push msg to dlq topic success")
					return
				}

				// TODO: consider push msg to retry topic
				consumerMsg.MetaData.CounterRetry = counterRetry + 1
				ttl := time.Duration(consumerMsg.MetaData.CounterRetry%maxRetry*5) * time.Second
				time.Sleep(ttl) // delay before resend retry topic
				newValue, _ := json.Marshal(consumerMsg)
				log.Printf("new value => %v \n", string(newValue))
				err = cgrh.consumerGroupHandler.retryPublisher.Publish(cgrh.consumerGroupHandler.ctx, nil, newValue)
				if err != nil {
					log.Printf("push msg to retry topic failed, error: %v \n", err)
				}
			}()
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
