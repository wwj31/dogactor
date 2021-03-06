<p align="center">game server framework</a></p>

a actor model framework written in Go.
it has implemented server discovery using ETCD.

Getting Started
===============

# 1.Installation
#### To start using dogactor, install Go and run `go get`:
```sh
$ go get -u github.com/wwj31/dogactor
```

# 2.Quick Start
Copy and paste the following code in your main files
```go
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
	ping := actor.New("ping", &PingActor{})
	pong := actor.New("pong", &PongActor{})
	system.Add(ping)
	system.Add(pong)

	<-system.CStop
	fmt.Println("stop")
}

// OnInit PingActor
func (s *PingActor) OnInit() {
	s.AddTimer("5h4j3kg4a3v9", tools.NowTime()+int64(1*time.Second), func(dt int64) {
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

// OnHandleMessage PongActor
func (s *PongActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch msg {
	case "this is data":
		fmt.Println(sourceId, targetId)
		s.Send(sourceId, 99999)
	}
}

