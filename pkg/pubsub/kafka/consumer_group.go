package kafka

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/tikivn/ims-library/pkg/library/log"
	"github.com/tikivn/ims-library/pkg/library/pubsub"
	"go.uber.org/zap"
)

// handlerFunc is a function which will be assign from our request. It's responsibility
// is handle message(s) had sent to topic by publisher
type handlerFunc func(msg *sarama.ConsumerMessage) error

// handlerFunc is a function which will be assign from our request. It's responsibility
// is handle message that an exception error has occurred when handlerFunc handling
type handlerFuncWithError func(msg *sarama.ConsumerMessage) error

type ConsumerGroupOption func(*consumerGroupSubscriber) error

type consumerGroupSubscriber struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	logger     *zap.Logger

	messageChan chan *consumerGroupMessage
	closeChan   chan bool

	messageRetryChan chan *consumerGroupMessage
	closeRetryChan   chan bool
	maxRetry         int
	baseTimeDelay    time.Duration

	err error

	brokers []string
	topic   string
	group   string

	handlerFunc          handlerFunc
	handlerFuncWithError handlerFuncWithError

	saramaConfig        *sarama.Config
	saramaConsumerGroup sarama.ConsumerGroup

	numberOfWorker int
}

func defaultConsumerGroup() (*consumerGroupSubscriber, error) {
	logger, err := log.DefaultConfig().Build()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_4_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	// we always want to see errors, no matter what
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Session.Timeout = 30 * time.Second

	return &consumerGroupSubscriber{
		ctx:            ctx,
		cancelFunc:     cancel,
		logger:         logger,
		err:            nil,
		brokers:        []string{"127.0.0.1:9092"},
		topic:          "default_topic",
		group:          "default_group_id",
		saramaConfig:   saramaConfig,
		numberOfWorker: 1,
	}, nil
}

func NewConsumerGroupSubscriber(opts ...ConsumerGroupOption) (pubsub.Subscriber, error) {
	cgs, err := defaultConsumerGroup()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		_ = opt(cgs)
	}
	defaultHandlerFuncWithError(cgs)

	cgs.saramaConsumerGroup, err = sarama.NewConsumerGroup(cgs.brokers, cgs.group, cgs.saramaConfig)
	if err != nil {
		return nil, err
	}
	return cgs, nil
}

func defaultHandlerFuncWithError(cgs *consumerGroupSubscriber) {
	if cgs.handlerFuncWithError == nil {
		cgs.handlerFuncWithError = func(msg *sarama.ConsumerMessage) error {
			topicDLQ := cgs.topic + "_DLQ"
			producer, err := NewPublisher(cgs.brokers)
			if err != nil {
				return err
			}
			err = producer.Publish(cgs.ctx, topicDLQ, msg.Value, WithPublisherMsgKey(msg.Key))
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func WithBrokers(brokers []string) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.brokers = brokers
		return nil
	}
}

func WithTopic(topic string) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.topic = topic
		return nil
	}
}

func WithGroupID(group string) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		if group == "" {
			return ErrConsumerGroupIsInvalid
		}
		cg.group = group
		return nil
	}
}

func WithHandlerFunc(handlerFunc func(message *sarama.ConsumerMessage) error) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		if handlerFunc == nil {
			return ErrHandlerFuncIsInvalid
		}
		cg.handlerFunc = handlerFunc
		return nil
	}
}

func WithHandlerWithErrorFunc(handlerWithErrorFunc func(message *sarama.ConsumerMessage) error) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.handlerFuncWithError = handlerWithErrorFunc
		return nil
	}
}

func WithSaramaConfig(cfg *sarama.Config) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.saramaConfig = cfg
		return nil
	}
}

func WithNumberOfWorker(number int) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		if number < 0 {
			return ErrNumberOfWorkerHandlerInvalid
		}
		cg.numberOfWorker = number
		return nil
	}
}

func WithOffsetsInitial(offset int64) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.saramaConfig.Consumer.Offsets.Initial = offset
		return nil
	}
}

func WithMaxRetry(maxRetry int) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.maxRetry = maxRetry
		return nil
	}
}

func WithBaseTimeDelay(baseTimeDelay time.Duration) ConsumerGroupOption {
	return func(cg *consumerGroupSubscriber) error {
		cg.baseTimeDelay = baseTimeDelay
		return nil
	}
}

func (cgs *consumerGroupSubscriber) Start() error {
	cgs.closeChan = make(chan bool)
	cgs.messageChan = make(chan *consumerGroupMessage)

	cgs.closeRetryChan = make(chan bool)
	cgs.messageRetryChan = make(chan *consumerGroupMessage)

	// initial consumerGroupHandler
	cgh := &consumerGroupHandler{
		ctx:         cgs.ctx,
		logger:      cgs.logger,
		readyChan:   make(chan bool),
		messageChan: cgs.messageChan,
		maxRetry:    cgs.maxRetry,
	}

	go func() {
		defer func() {
			cgs.logger.Info("close messageChan channel")
			close(cgs.messageChan)
			cgs.logger.Info("close messageRetryChan channel")
			close(cgs.messageRetryChan)
		}()

		for {
			if err := cgs.saramaConsumerGroup.Consume(cgs.ctx, []string{cgs.topic}, cgh); err != nil {
				cgs.logger.Error("Exception error has occurred from consumer",
					zap.Any("error", err),
				)
				cgs.err = err
				close(cgh.readyChan)
				return
			}

			// check if context was cancelled, signaling that the consumer should stop
			if err := cgs.ctx.Err(); err != nil {
				return
			}
		}
	}()

	// Await till the consumer has been set up
	// readyChan channel will be received signal fro consumer when consumer has set up
	// We can check in `Setup` function (method) of consumerGroupHandler
	<-cgh.readyChan
	cgs.logger.Info("Consumer up and running...",
		zap.Any("brokers", cgs.brokers),
		zap.Any("topic", cgs.topic),
		zap.Any("group", cgs.group),
	)

	// consume message from main_message channel which contains message from kafka topic
	go cgs.consume(cgs.messageChan, cgs.closeChan)

	// consume message from retry_message channel which contains message failed when we
	// handle in main_message channel
	go cgs.consume(cgs.messageRetryChan, cgs.closeRetryChan)

	// create a channel sigterm to listen os signal. If service is killed or crash, we will release
	// all resource or goroutine(s)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-cgs.ctx.Done():
		cgs.logger.Info("terminating: context cancelled")
	case <-sigterm:
		cgs.logger.Info("terminating: via signal")
	}

	if err := cgs.Stop(); err != nil {
		cgs.logger.Error("Cannot stop subscriber: ", zap.Any("error", err))
	}

	// wait for subscriber to stop
	<-cgs.closeChan
	// wait for subscriber retry to stop
	<-cgs.closeRetryChan

	cgs.logger.Info("consumer group subscriber stopped!")
	return nil
}

func (cgs *consumerGroupSubscriber) consume(messageChan <-chan *consumerGroupMessage, closeChan chan bool) {
	defer close(closeChan)

	wg := &sync.WaitGroup{}
	wg.Add(cgs.numberOfWorker)
	for i := 0; i < cgs.numberOfWorker; i++ {
		go func() {
			defer wg.Done()

			for {
				msg, ok := <-messageChan
				if !ok {
					return
				}

				if msg.message.Value == nil {
					cgs.logger.Warn("content message is nil")
					continue
				}

				// TODO: please find the better solution for delay time before retry
				if msg.counterRetry > 0 {
					timeDelay := time.Duration(msg.counterRetry) * cgs.baseTimeDelay
					time.Sleep(timeDelay)
				}

				if err := cgs.handlerFunc(msg.message); err != nil {
					cgs.logger.Error("consume message error",
						zap.Any("topic", msg.message.Topic),
						zap.Any("partition", msg.message.Partition),
						zap.Any("offset", msg.message.Offset),
						zap.Any("error", err),
					)

					// push message into retry_channel to retry message
					if msg.counterRetry < msg.maxRetry {
						msg.counterRetry++
						cgs.messageRetryChan <- msg
						continue
					}

					// if we have tried it again and again and failed, push message to DLQ is good option
					if cgs.handlerFuncWithError != nil {
						cgs.handlerFuncWithError(msg.message)
					}
				}

				if err := msg.Done(); err != nil {
					cgs.logger.Error("ACK message failed",
						zap.Any("topic", msg.message.Topic),
						zap.Any("partition", msg.message.Partition),
						zap.Any("offset", msg.message.Offset),
						zap.Any("error", err),
					)
				}
			}
		}()
	}
	wg.Wait()
}

func (cgs *consumerGroupSubscriber) Error() error {
	return cgs.err
}

func (cgs *consumerGroupSubscriber) Stop() error {
	cgs.cancelFunc()
	cgs.logger.Info("context cancel function!")

	if err := cgs.saramaConsumerGroup.Close(); err != nil {
		cgs.logger.Error("Error closing consumer group",
			zap.Any("error", err),
		)
		return err
	}
	cgs.logger.Info("sarama consumer group close!")
	return nil
}
