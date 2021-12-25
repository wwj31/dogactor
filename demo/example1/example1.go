package main

import (
	"fmt"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cmd"
	"github.com/wwj31/dogactor/tools"
	"time"
)

type PingActor struct{ actor.Base }
type PongActor struct{ actor.Base }

func main() {
	system, _ := actor.NewSystem(actor.WithCMD(cmd.New()))
	ping := actor.New("ping", &PingActor{}, actor.SetMailBoxSize(5000))
	pong := actor.New("pong", &PongActor{}, actor.SetMailBoxSize(5000))
	system.Add(ping)
	system.Add(pong)

	<-system.CStop

	fmt.Println("stop")
}

// PingActor
func (s *PingActor) OnInit() {
	s.AddTimer(tools.UUID(), tools.NowTime()+int64(1*time.Second), func(dt int64) {
		s.Send("pong", "this is data")
	}, -1)
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
		//fmt.Println(sourceId, targetId)
		s.Send(sourceId, 99999)
	}
}
