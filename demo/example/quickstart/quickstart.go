package main

import (
	"fmt"
	"time"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
)

type PingActor struct{ actor.Base }
type PongActor struct{ actor.Base }

func main() {
	system, _ := actor.NewSystem()
	system.NewActor("ping", &PingActor{})
	system.NewActor("pong", &PongActor{})

	<-system.Stopped

	fmt.Println("stop")
}

// OnInit PingActor
func (s *PingActor) OnInit() {
	s.AddTimer(tools.XUID(), tools.Now().Add(time.Second), func(dt time.Duration) {
		s.Send("pong", "this is a msg from pong")
	}, -1)
}
func (s *PingActor) OnHandle(msg actor.Message) {
	switch msg.RawMsg() {
	case 99999:
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		fmt.Println()
	}
}

// OnHandle PongActor
func (s *PongActor) OnHandle(msg actor.Message) {
	switch msg.RawMsg() {
	case "this is a msg from pong":
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		s.Send(msg.GetSourceId(), 99999)
	}
}
