package test

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestPull(t *testing.T) {
	// Use the env variable if running in the container, otherwise use the default.
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	// Create an unauthenticated connection to NATS.
	nc, _ := nats.Connect(url)
	defer nc.Drain()

	// Access `JetStreamContext` to use the JS APIs.
	js, _ := nc.JetStream()

	fmt.Println("streams:")
	for s := range js.Streams() {
		fmt.Printf("%+v", s)
	}

	// ### Creating the stream
	// Define the stream configuration, specifying `WorkQueuePolicy` for
	// retention, and create the stream.
	cfg := &nats.StreamConfig{
		Name:      "pull_test",
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{"actor1.>"},
		Storage:   nats.MemoryStorage,
		NoAck:     false,
	}

	if _, err := js.AddStream(cfg); err != nil {
		if jsErr, ok := err.(nats.JetStreamError); !ok || jsErr.APIError().ErrorCode != nats.JSErrCodeStreamNameInUse {
			assert.NoError(t, err)
			return
		}
	}

	_, updateerr := js.UpdateStream(cfg)
	assert.NoError(t, updateerr)
	fmt.Println("created the stream")

	// Checking the stream info, we see three messages have been queued.
	sub1, _ := js.PullSubscribe("", "processor-1", nats.BindStream(cfg.Name))
	//nats.PullMaxWaiting(3), // 最大等待的fetch连接数量

	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(time.Second)
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(time.Second)
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(time.Second)
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(time.Second)

	g := func(i int) {
		//time.Sleep(10 * time.Second)
		fmt.Println("# Stream info without any consumers")

		for {
			//printStreamState(js, cfg.Name)
			t1 := time.Now()
			fmt.Println("wait fetch ", i)
			msgs, fetErr := sub1.Fetch(10, nats.MaxWait(10*time.Second))
			assert.NoError(t, fetErr)
			fmt.Println("delay ", time.Now().Sub(t1).String())
			var lastMsg *nats.Msg
			for _, msg := range msgs {
				msg.Ack()
				fmt.Println(i, string(msg.Data))
				lastMsg = msg
			}
			if lastMsg != nil {
				//assert.NoError(t, lastMsg.Ack())
			}
			//time.Sleep(1 * time.Second)
		}
	}
	go g(1)
	//go g(2)

	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//for {
	//	js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//	time.Sleep(100 * time.Millisecond)
	//
	//}
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(1 * time.Second)
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(1 * time.Second)
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(1 * time.Second)
	//js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//time.Sleep(1 * time.Second)
	//go func() {
	//	for {
	//		randInt := time.Duration(1)
	//		time.Sleep(randInt * time.Second)
	//		js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//		//fmt.Println("push ", randInt)
	//	}
	//}()

	select {}
}
