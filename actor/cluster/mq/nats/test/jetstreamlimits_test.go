package test

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestJetStreamLimits(t *testing.T) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	assert.NoError(t, err)

	defer nc.Drain()

	js, jerr := nc.JetStream()
	assert.NoError(t, jerr)

	cfg := nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"events.>"},
	}

	cfg.Storage = nats.FileStorage

	js.AddStream(&cfg)
	fmt.Println("created the stream")

	js.Publish("events.page_loaded", nil)
	js.Publish("events.mouse_clicked", nil)
	js.Publish("events.mouse_clicked", nil)
	js.Publish("events.page_loaded", nil)
	js.Publish("events.mouse_clicked", nil)
	js.Publish("events.input_focused", nil)
	fmt.Println("published 6 messages")

	js.PublishAsync("events.input_changed", nil)
	js.PublishAsync("events.input_blurred", nil)
	js.PublishAsync("events.key_pressed", nil)
	js.PublishAsync("events.input_focused", nil)
	js.PublishAsync("events.input_changed", nil)
	js.PublishAsync("events.input_blurred", nil)

	select {
	case <-js.PublishAsyncComplete():
		fmt.Println("published 6 messages")
	case <-time.After(time.Second):
		log.Fatal("publish took too long")
	}

	printStreamState(js, cfg.Name)

	cfg.MaxMsgs = 10
	js.UpdateStream(&cfg)
	fmt.Println("set max messages to 10")

	printStreamState(js, cfg.Name)

	cfg.MaxBytes = 300
	js.UpdateStream(&cfg)
	fmt.Println("set max bytes to 300")

	printStreamState(js, cfg.Name)

	cfg.MaxAge = time.Second
	js.UpdateStream(&cfg)
	fmt.Println("set max age to one second")

	printStreamState(js, cfg.Name)

	fmt.Println("sleeping one second...")
	time.Sleep(time.Second)

	printStreamState(js, cfg.Name)
}

func printStreamState(js nats.JetStreamContext, name string) {
	info, _ := js.StreamInfo(name)
	b, _ := json.MarshalIndent(info.State, "", " ")
	fmt.Println("inspecting stream info")
	fmt.Println(string(b))
}
