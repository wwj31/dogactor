package main

import (
	"fmt"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
	"time"
)

type PingActor struct{ actor.Base }
type PongActor struct{ actor.Base }

func main() {
	system, _ := actor.NewSystem()
	ping := actor.New("ping", &PingActor{}, actor.SetMailBoxSize(500))
	pong := actor.New("pong", &PongActor{}, actor.SetMailBoxSize(500))
	system.Add(ping)
	system.Add(pong)

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
	switch msg.Message() {
	case 99999:
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		fmt.Println()
	}
}

// OnHandleMessage PongActor
func (s *PongActor) OnHandle(msg actor.Message) {
	switch msg.Message() {
	case "this is data":
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		s.Send(msg.GetSourceId(), 99999)
	}
}
