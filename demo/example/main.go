package main

import (
	"fmt"
	"time"

	"github.com/wwj31/godactor/actor"
)

//
//var (
//	ETCD_ADDR   = "127.0.0.1:2379"
//	ETCD_PREFIX = "demo/"
//)

type PingActor struct{ actor.ActorHanlerBase }
type PongActor struct{ actor.ActorHanlerBase }

func main() {
	system, _ := actor.System()
	ping := actor.New("ping", &PingActor{})
	pong := actor.New("pong", &PongActor{})
	system.Regist(ping)
	system.Regist(pong)
	select {}
}

// PingActor
func (s *PingActor) Init() {
	s.AddTimer(2*time.Second, -1, func(dt int64) {
		s.Send("pong", "this is data")
	})
}
func (s *PingActor) HandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case 99999:
		fmt.Println(sourceId, targetId)
		fmt.Println()
	}
}

//PongActor
func (s *PongActor) HandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case "this is data":
		fmt.Println(sourceId, targetId)
		s.Send(sourceId, 99999)
	}
}
