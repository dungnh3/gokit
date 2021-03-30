package main

import (
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/tikivn/ims-library/pkg/library/pubsub/kafka"
	"time"
)

const (
	Topic = "default_topic"
	Group = "test.group.105"
)

func main() {
	subscriber, err := kafka.NewConsumerGroupSubscriber(
		kafka.WithBrokers([]string{"127.0.0.1:9092"}),
		kafka.WithTopic(Topic),
		kafka.WithGroupID(Group),
		kafka.WithOffsetsInitial(sarama.OffsetNewest),
		kafka.WithNumberOfWorker(4),
		//kafka.WithHandlerFunc(showMessage),
		kafka.WithHandlerFunc(testRetryMessage),
		kafka.WithMaxRetry(3),
		kafka.WithBaseTimeDelay(5*time.Second),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("start consumer group..")
	subscriber.Start()
}

func showMessage(msg *sarama.ConsumerMessage) error {
	fmt.Println("wait 5s")
	time.Sleep(10 * time.Second)
	fmt.Println(string(msg.Value))
	fmt.Println(fmt.Sprintf("partirion [%v], offset [%v]", msg.Partition, msg.Offset))
	return nil
}

func testRetryMessage(msg *sarama.ConsumerMessage) error {
	fmt.Println(string(msg.Value))
	fmt.Println(fmt.Sprintf("partirion [%v], offset [%v]", msg.Partition, msg.Offset))
	return errors.New("fake error")
}
