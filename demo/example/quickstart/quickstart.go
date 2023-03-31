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
	system.NewActor("ping", &PingActor{}, actor.SetMailBoxSize(500))
	system.NewActor("pong", &PongActor{}, actor.SetMailBoxSize(500))

	<-system.Stopped

	fmt.Println("stop")
}

// OnInit PingActor
func (s *PingActor) OnInit() {
	s.AddTimer(tools.XUID(), tools.Now().Add(time.Second), func(dt time.Duration) {
		s.Send("pong", "this is data")
	}, -1)
}
func (s *PingActor) OnHandle(msg actor.Message) {
	switch msg.RawMsg() {
	case 99999:
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		fmt.Println()
	}
}

// OnHandleMessage PongActor
func (s *PongActor) OnHandle(msg actor.Message) {
	switch msg.RawMsg() {
	case "this is data":
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		s.Send(msg.GetSourceId(), 99999)
	}
}
