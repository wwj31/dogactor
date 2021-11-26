package main

import (
	"fmt"
	"github.com/wwj31/dogactor/actor/cmd"
	"github.com/wwj31/dogactor/log"
	"time"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
)

type PingActor struct{ actor.Base }
type PongActor struct{ actor.Base }

func main() {
	log.Init(log.TAG_DEBUG_I, nil, "./_log", "demo", 1)
	system, _ := actor.NewSystem(actor.WithCMD(cmd.New()))
	ping := actor.New("ping", &PingActor{}, actor.SetMailBoxSize(5000))
	pong := actor.New("pong", &PongActor{}, actor.SetMailBoxSize(5000))
	system.Regist(ping)
	system.Regist(pong)

	<-system.CStop
	log.Stop()
	fmt.Println("stop")
}

// PingActor
func (s *PingActor) OnInit() {
	s.AddTimer(tools.UUID(), 1*time.Second, func(dt int64) {
		for i := 0; i < 1000; i++ {
			s.Send("pong", "this is data")
		}
		tools.PrintMemUsage()
	}, -1)

	//s.AddTimer(tools.UUID(), 10*time.Second, func(dt int64) {
	//	s.System().Stop()
	//}, 1)
}
func (s *PingActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case 99999:
		//fmt.Println(sourceId, targetId)
		//fmt.Println()
	}
}

//PongActor
func (s *PongActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case "this is data":
		//fmt.Println(sourceId, targetId)
		s.Send(sourceId, 99999)
	}
}
