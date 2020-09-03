package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type consumerGroup struct {
	ctx                        context.Context
	topic                      string
	name                       string
	retryProducer              *producer
	dlqProducer                *producer
	saramaConsumerGroup        sarama.ConsumerGroup
	saramaConsumerGroupHandler sarama.ConsumerGroupHandler
}

func (cg *consumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	go func() {
		for {
			if err := cg.saramaConsumerGroup.Consume(ctx, topics, handler); err != nil {
				log.Panicf("error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}

func (cg *consumerGroup) Errors() <-chan error {
	return nil
}

func (cg *consumerGroup) Close() []error {
	var errs []error
	errs = append(errs, cg.retryProducer.Close(), cg.dlqProducer.Close(), cg.saramaConsumerGroup.Close())
	return errs
}

func (cg *consumerGroup) Start() error {
	return cg.Consume(cg.ctx, strings.Split(cg.topic, ","), cg.saramaConsumerGroupHandler)
}

func newConsumerGroup(ctx context.Context, cfg *sarama.Config, brokers []string, topic string, groupID string, handler sarama.ConsumerGroupHandler) (*consumerGroup, error) {
	var err error
	cg := new(consumerGroup)
	cg.ctx = ctx
	cg.topic = topic
	cg.name = "consumer_" + topic
	cg.saramaConsumerGroupHandler = handler
	cg.saramaConsumerGroup, err = sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return cg, nil
}

type kafkaConsumerGroup struct {
	ctx            context.Context
	cancel         context.CancelFunc
	consumerGroups []*consumerGroup
}

func NewKafkaConsumerGroup(cfg *ConsumerConfig, groupID string, handler func(msg *sarama.ConsumerMessage) error) (*kafkaConsumerGroup, error) {
	var err error
	kafkaConsumerGroup := new(kafkaConsumerGroup)

	if cfg.saramaConfig == nil {
		cfg.saramaConfig = sarama.NewConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	kafkaConsumerGroup.ctx = ctx
	kafkaConsumerGroup.cancel = cancel

	retryTopic := cfg.Topic + "_retry"
	dlqTopic := cfg.Topic + "_dlq"

	// TODO: create retry publisher to push msg into retry topic
	retryPublisher, err := NewPublisher(&ProducerConfig{
		Brokers:      cfg.BrokersRetry,
		Topic:        retryTopic,
		saramaConfig: cfg.saramaConfig,
	})
	if err != nil {
		return nil, err
	}

	// TODO: create dlq publisher to push msg into dlq topic
	dlqPublisher, err := NewPublisher(&ProducerConfig{
		Brokers:      cfg.BrokersDLQ,
		Topic:        dlqTopic,
		saramaConfig: cfg.saramaConfig,
	})
	if err != nil {
		return nil, err
	}

	// TODO: create new consumer group
	consumerGroupHandler := &consumerGroupHandler{
		ctx:           ctx,
		maxRetry:      cfg.MaxRetry,
		retryProducer: retryPublisher,
		dlqProducer:   dlqPublisher,
		handler:       handler,
	}
	consumerGroup, err := newConsumerGroup(ctx, cfg.saramaConfig, cfg.Brokers, cfg.Topic, groupID, consumerGroupHandler)
	if err != nil {
		return nil, err
	}

	// TODO: create new consumer retry group
	consumerGroupRetryHandler := &consumerGroupRetryHandler{
		consumerGroupHandler: consumerGroupHandler,
	}
	consumerGroupRetry, err := newConsumerGroup(ctx, cfg.saramaConfig, cfg.BrokersRetry, retryTopic, groupID+"_retry", consumerGroupRetryHandler)
	if err != nil {
		return nil, err
	}
	kafkaConsumerGroup.consumerGroups = append(kafkaConsumerGroup.consumerGroups, consumerGroup, consumerGroupRetry)
	return kafkaConsumerGroup, nil
}

func (kcg *kafkaConsumerGroup) Start() {
	ready := make(chan bool)
	go func(ready chan bool) {
		for index, _ := range kcg.consumerGroups {
			err := kcg.consumerGroups[index].Start()
			if err != nil {
				log.Printf("start consumer group failed, error %v \n", err)
				ready <- false
			}
			log.Printf("%v is running...\n", kcg.consumerGroups[index].name)
		}
	}(ready)
	log.Println("kafka consumer group is running...")
	ctx := kcg.ctx
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	case <-ready:
		log.Println("terminating: start consumer failed")
	}
	kcg.cancel()
	if errs := kcg.Stop(); len(errs) != 0 {
		log.Panicf("error closing client: %v", errs)
	}
	return
}

func (kcg *kafkaConsumerGroup) Stop() []error {
	var errs []error
	for index, _ := range kcg.consumerGroups {
		kerrs := kcg.consumerGroups[index].Close()
		if kerrs != nil {
			log.Printf("stop consumer group failed, error %v \n", kerrs)
			errs = append(errs, kerrs...)
		}
	}
	return errs
}
