package mq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type MQConsumer struct{}

func Start() {
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("rights_change_consumer_group"),
	)

	c.Subscribe("topicName", consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			fmt.Println(string(msg.Body))
		}
		return consumer.ConsumeSuccess, nil
	})

	err := c.Start()
	if err != nil {
		fmt.Println(err)
	}
}
