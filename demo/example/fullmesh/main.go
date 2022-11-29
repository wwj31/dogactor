package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh"
	"github.com/wwj31/dogactor/demo/example/fullmesh/msg"
	"github.com/wwj31/dogactor/logger"
	"github.com/wwj31/dogactor/tools"
)

var (
	ETCD_ADDR   = "127.0.0.1:2379"
	ETCD_PREFIX = "dog"
)

type Student struct {
	actor.Base
	Name string
	Age  int
}

var log *logger.Logger

func main() {
	log = logger.New(logger.Option{
		Level:          logger.DebugLevel,
		LogPath:        "./example2",
		FileName:       "e.log",
		FileMaxAge:     1,
		FileMaxSize:    100,
		FileMaxBackups: 1,
		DisplayConsole: true,
		Skip:           1,
	})
	protoIndex := tools.NewProtoIndex(func(name string) (interface{}, bool) {
		return msg.Spawner(name)
	}, tools.EnumIdx{})

	system1, _ := actor.NewSystem(actor.Addr("127.0.0.1:5000"),
		fullmesh.WithRemote(ETCD_ADDR, ETCD_PREFIX), actor.ProtoIndex(protoIndex))
	system2, _ := actor.NewSystem(actor.Addr("127.0.0.1:5001"),
		fullmesh.WithRemote(ETCD_ADDR, ETCD_PREFIX), actor.ProtoIndex(protoIndex))

	lilei := actor.New("LiLei", &Student{Name: "LiLei", Age: 19})
	hanmeimei := actor.New("HanMeimei", &Student{Name: "HanMeimei", Age: 15})

	system1.Add(lilei)
	system2.Add(hanmeimei)

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
		log.Infof(m.Data)

		reqId := actor.RequestId(v.GetRequestId())
		if reqId.Valid() {
			s.Response(reqId, &msg.HanMeimeiSay{
				Data: "no!",
			})
		} else {
			s.Send(v.GetSourceId(), &msg.HanMeimeiSay{
				Data: "hi~! Li Lei, I'm HanMeimei",
			})
		}

	case *msg.HanMeimeiSay:
		log.Infof(m.Data)

		resp, _ := s.RequestWait(v.GetSourceId(), &msg.LileiSay{
			Data: "Be my grilfriend?",
		})
		// waiting....
		log.Infof(resp.(*msg.HanMeimeiSay).Data)

		s.Request(v.GetSourceId(), &msg.LileiSay{
			Data: "it's ok! I will protect you.",
		}).Handle(func(resp interface{}, err error) {
			log.Infof(resp.(*msg.HanMeimeiSay).Data)
		})

		s.AddTimer(tools.XUID(), tools.Now().Add(time.Second), func(dt time.Duration) {
			s.Request("HanMeimei", &msg.LileiSay{
				Data: "please~",
			}).Handle(func(resp interface{}, err error) {
				log.Infof(resp.(*msg.HanMeimeiSay).Data)
			})
		}, -1)
	}
}
