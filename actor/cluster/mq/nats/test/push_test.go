package test

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestPush(t *testing.T) {
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
		Name:      "pull_test2",
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{"actor2.>"},
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
	_, _ = js.Subscribe("", func(msg *nats.Msg) {
		fmt.Println("msg:", string(msg.Data))

	}, nats.BindStream(cfg.Name))

	for {
		js.PublishAsync("actor2.msg", []byte(time.Now().String()))
		time.Sleep(1 * time.Second)

	}
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
	//		randInt := time.Duration(rand.Intn(5))
	//		time.Sleep(randInt * time.Second)
	//		js.PublishAsync("actor1.msg", []byte(time.Now().String()))
	//		fmt.Println("push ", randInt)
	//	}
	//}()

	select {}
}
