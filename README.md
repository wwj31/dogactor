[![Go Report Card](https://goreportcard.com/badge/github.com/wwj31/dogactor)](https://goreportcard.com/report/github.com/wwj31/dogactor)
![GitHub closed pull requests](https://img.shields.io/github/issues-pr-closed-raw/wwj31/dogactor)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/wwj31/dogactor?color=red&label=Go)

<p align="center">game server framework</a></p>


a actor model framework written in Go, two ways to communicate internally： 
1. Service discovery based on ETCD.
2. Use NATS(jetStream) as a means of communication(default).
# <img align="center" src="https://github.com/wwj31/dogactor/raw/master/.github/images/image.png" alt="img" title="img" />
# 1.Installation
#### To start using dogactor, install Go and run `go get`:
```sh
$ go get -u github.com/wwj31/dogactor
```

# 2.Quick Start
Code below is a simple implementation of dogactor,
run it after copy to your main, or see [demo/example](demo/example)
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
	switch msg.Payload() {
	case 99999:
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
	}
}

// OnHandle PongActor
func (s *PongActor) OnHandle(msg actor.Message) {
	switch msg.Payload() {
	case "this is a msg from pong":
		fmt.Println(msg.GetSourceId(), msg.GetTargetId())
		s.Send(msg.GetSourceId(), 99999)
	}
}
```
