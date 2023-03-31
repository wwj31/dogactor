package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"github.com/wwj31/dogactor/actor/cluster/mq/nats"
	"github.com/wwj31/dogactor/demo/example/fullmesh/msg"
	"github.com/wwj31/dogactor/logger"
	"github.com/wwj31/dogactor/tools"
)

var (
	natsUrl = "nats://localhost:4222"
)

type Student struct {
	actor.Base
	Name string
	Age  int
}

func main() {
	protoIndex := tools.NewProtoIndex(func(name string) (interface{}, bool) {
		return msg.Spawner(name)
	}, tools.EnumIdx{})

	system1, _ := actor.NewSystem(mq.WithRemote(natsUrl, nats.New()), actor.ProtoIndex(protoIndex))
	system2, _ := actor.NewSystem(mq.WithRemote(natsUrl, nats.New()), actor.ProtoIndex(protoIndex))
	system1.NewActor("LiLei", &Student{Name: "LiLei", Age: 19})
	system2.NewActor("HanMeimei", &Student{Name: "HanMeimei", Age: 15})

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	system1.Stop()
	system2.Stop()
	<-system1.Stopped
	<-system2.Stopped

	logger.Close()
}

func (s *Student) OnInit() {
	if s.Name == "LiLei" {
		s.AddTimer(tools.XUID(), tools.Now().Add(2*time.Second), func(dt time.Duration) {
			s.Send("HanMeimei", &msg.LileiSay{Data: "hello, I'm Li Lei"})
		})
	}
}

func (s *Student) OnHandle(v actor.Message) {
	switch m := v.RawMsg().(type) {
	case *msg.LileiSay:
		fmt.Println(m.Data)

		if v.GetRequestId() != "" {
			s.Response(v.GetRequestId(), &msg.HanMeimeiSay{
				Data: "no!",
			})
		} else {
			s.Send(v.GetSourceId(), &msg.HanMeimeiSay{
				Data: "hi~! Li Lei, I'm HanMeimei",
			})
		}

	case *msg.HanMeimeiSay:
		fmt.Println(m.Data)

		resp, _ := s.RequestWait(v.GetSourceId(), &msg.LileiSay{
			Data: "Be my grilfriend?",
		})
		// waiting....
		fmt.Println(resp.(*msg.HanMeimeiSay).Data)

		s.Request(v.GetSourceId(), &msg.LileiSay{
			Data: "it's ok! I will protect you.",
		}).Handle(func(resp interface{}, err error) {
			fmt.Println(resp.(*msg.HanMeimeiSay).Data)
		})

		s.AddTimer(tools.XUID(), tools.Now().Add(time.Second), func(dt time.Duration) {
			resp, _ := s.RequestWait("HanMeimei", &msg.LileiSay{
				Data: "please~",
			})
			fmt.Println(resp.(*msg.HanMeimeiSay).Data)

		}, -1)
	}
}
