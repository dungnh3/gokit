package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tikivn/ims-library/pkg/library/pubsub/kafka"
)

func main() {
	ctx := context.Background()
	brokers := []string{"127.0.0.1:9092"}
	producer, err := kafka.NewPublisher(brokers)
	if err != nil {
		panic(err)
	}
	msg := "hello world - demo - 123"
	msgValue, _ := json.Marshal(msg)

	for i := 0; i < 1; i++ {
		err := producer.Publish(ctx,
			"default_topic",
			/* message */ msgValue,
			kafka.WithPublisherMsgPartition(int32(i)),
			kafka.WithPublisherMsgHeader(map[string]string{
				"key":  "value",
				"key2": "value2",
			}),
		)
		if err != nil {
			fmt.Errorf("publish message to kafka broker failed with error: %v", err.Error())
			return
		}
	}

	fmt.Println("Done")
}
