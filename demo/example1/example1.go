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
	//log.Init(log.TAG_DEBUG_I, nil, "./_log", "demo", 1)
	system, _ := actor.NewSystem()
	ping := actor.New("ping", &PingActor{})
	pong := actor.New("pong", &PongActor{})
	system.Regist(ping)
	system.Regist(pong)

	<-system.CStop
	fmt.Println("stop")
}

// PingActor
func (s *PingActor) OnInit() {
	s.AddTimer(tools.UUID(), 2*time.Second, func(dt int64) {
		s.Send("pong", "this is data")
	}, -1)

	s.AddTimer(tools.UUID(), 10*time.Second, func(dt int64) {
		s.System().Stop()
	}, 1)
}
func (s *PingActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case 99999:
		fmt.Println(sourceId, targetId)
		fmt.Println()
	}
}

//PongActor
func (s *PongActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case "this is data":
		fmt.Println(sourceId, targetId)
		s.Send(sourceId, 99999)
	}
}
