package test

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	js, _ := nc.JetStream()

	cfg := &nats.StreamConfig{
		Name:      "EVENTS",
		Retention: nats.InterestPolicy,
		Subjects:  []string{"events.>"},
	}

	js.AddStream(cfg)
	fmt.Println("created the stream")

	js.Publish("events.page_loaded", nil)
	js.Publish("events.mouse_clicked", nil)
	ack, _ := js.Publish("events.input_focused", nil)
	fmt.Println("published 3 messages")

	fmt.Printf("last message seq: %d\n", ack.Sequence)

	fmt.Println("# Stream info without any consumers")
	printStreamState(js, cfg.Name)

	js.AddConsumer(cfg.Name, &nats.ConsumerConfig{
		Durable:   "processor-1",
		AckPolicy: nats.AckExplicitPolicy,
	})

	js.Publish("events.mouse_clicked", nil)
	js.Publish("events.input_focused", nil)

	fmt.Println("\n# Stream info with one consumer")
	printStreamState(js, cfg.Name)

	sub1, _ := js.PullSubscribe("", "processor-1", nats.Bind(cfg.Name, "processor-1"))
	defer sub1.Unsubscribe()

	msgs, _ := sub1.Fetch(2)
	msgs[0].Ack()
	msgs[1].AckSync()

	fmt.Println("\n# Stream info with one consumer and acked messages")
	printStreamState(js, cfg.Name)

	js.AddConsumer(cfg.Name, &nats.ConsumerConfig{
		Durable:   "processor-2",
		AckPolicy: nats.AckExplicitPolicy,
	})

	js.Publish("events.input_focused", nil)
	js.Publish("events.mouse_clicked", nil)

	sub2, _ := js.PullSubscribe("", "processor-2", nats.Bind(cfg.Name, "processor-2"))
	defer sub2.Unsubscribe()

	msgs, _ = sub2.Fetch(2)
	md0, _ := msgs[0].Metadata()
	md1, _ := msgs[1].Metadata()
	fmt.Printf("msg seqs %d and %d", md0.Sequence.Stream, md1.Sequence.Stream)
	msgs[0].Ack()
	msgs[1].AckSync()

	fmt.Println("\n# Stream info with two consumers, but only one set of acked messages")
	printStreamState(js, cfg.Name)

	msgs, _ = sub1.Fetch(2)
	msgs[0].Ack()
	msgs[1].AckSync()

	fmt.Println("\n# Stream info with two consumers having both acked")
	printStreamState(js, cfg.Name)

	js.AddConsumer(cfg.Name, &nats.ConsumerConfig{
		Durable:       "processor-3",
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: "events.mouse_clicked",
	})

	js.Publish("events.input_focused", nil)

	msgs, _ = sub1.Fetch(1)
	msgs[0].Term()
	msgs, _ = sub2.Fetch(1)
	msgs[0].AckSync()

	fmt.Println("\n# Stream info with three consumers with interest from two")
	printStreamState(js, cfg.Name)
}
