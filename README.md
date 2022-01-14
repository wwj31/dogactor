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
)
type PingActor struct{ actor.Base }
type PongActor struct{ actor.Base }

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
