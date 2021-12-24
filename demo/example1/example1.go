package main

import (
	"fmt"
	"github.com/wwj31/dogactor/actor/cmd"
	"github.com/wwj31/dogactor/l"
	"time"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/tools"
)

type PingActor struct{ actor.Base }
type PongActor struct{ actor.Base }

func main() {
	l.Init(l.Option{
		Level:          l.DebugLevel,
		LogPath:        "./example",
		FileName:       "e.log",
		FileMaxAge:     1,
		FileMaxSize:    100,
		FileMaxBackups: 1,
		DisplayConsole: true,
	})
	l.Infow("example start")
	system, _ := actor.NewSystem(actor.WithCMD(cmd.New()))
	ping := actor.New("ping", &PingActor{}, actor.SetMailBoxSize(5000))
	pong := actor.New("pong", &PongActor{}, actor.SetMailBoxSize(5000))
	system.Regist(ping)
	system.Regist(pong)

	<-system.CStop

	l.Close()
	fmt.Println("stop")
}

// PingActor
func (s *PingActor) OnInit() {
	s.AddTimer(tools.UUID(), 100*time.Millisecond, func(dt int64) {
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
