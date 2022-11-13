package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	nats2 "github.com/wwj31/dogactor/actor/cluster/mq/nats"
	"testing"
	"time"
)

// nats 是 at most once 投递消息，测试证明了没有消费者时，消息不会被保存
// 所以没有消息积压的能力
func TestPublishWithDelaySubscription(t *testing.T) {
	nats := nats2.New()
	assert.NoError(t, nats.Connect("nats://localhost:4222"))

	var sub = "topicA"
	go func() {
		for i := 1; i <= 5; i++ {
			time.Sleep(time.Second)
			fmt.Println("sleep ", i)
		}

		assert.NoError(t, nats.SubASync(sub, func(data []byte) {
			fmt.Println("subscribe ", string(data))
		}))
	}()

	for i := 0; i < 100000; i++ {
		assert.NoError(t, nats.Pub(sub, []byte(fmt.Sprintf("data:%v", i))))
		assert.NoError(t, nats.Flush())
		time.Sleep(time.Second)
	}

	select {}
}
