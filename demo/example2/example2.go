package main

import (
	"github.com/wwj31/dogactor/actor/cluster"
	"github.com/wwj31/dogactor/actor/cmd"
	"github.com/wwj31/dogactor/demo/example2/interval"
	"github.com/wwj31/dogactor/l"
	"github.com/wwj31/dogactor/tools"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwj31/dogactor/actor"
)

var (
	ETCD_ADDR   = "127.0.0.1:2379"
	ETCD_PREFIX = "demo/"
)

type Student struct {
	actor.Base
	Name string
	Age  int
}

var log *l.Logger

func main() {
	log = l.New(l.Option{
		Level:          l.DebugLevel,
		LogPath:        "./example2",
		FileName:       "e.log",
		FileMaxAge:     1,
		FileMaxSize:    100,
		FileMaxBackups: 1,
		DisplayConsole: true,
	})
	system1, _ := actor.NewSystem(actor.Addr("127.0.0.1:5000"), cluster.WithRemote(ETCD_ADDR, ETCD_PREFIX), actor.WithCMD(cmd.New()))
	lilei := actor.New("LiLei", &Student{Name: "LiLei", Age: 19})
	system1.Add(lilei)

	system2, _ := actor.NewSystem(actor.Addr("127.0.0.1:5001"), cluster.WithRemote(ETCD_ADDR, ETCD_PREFIX))
	hanmeimei := actor.New("HanMeimei", &Student{Name: "HanMeimei", Age: 15})
	system2.Add(hanmeimei)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	system1.Stop()
	system2.Stop()
	<-system1.CStop
	<-system2.CStop

	l.Close()
}
func (s *Student) OnInit() {
	if s.Name == "LiLei" {
		s.AddTimer(tools.UUID(), tools.NowTime()+int64(2*time.Second), func(dt int64) {
			s.Send("HanMeimei", &interval.LileiSay{Data: "hello, I'm Li Lei"})
		})
	}
}

func (s *Student) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch m := msg.(type) {
	case *interval.LileiSay:
		log.Infof(m.Data)

		s.Send(sourceId, &interval.HanMeimeiSay{
			Data: "hi~! Li Lei, I'm HanMeimei",
		})
	case *interval.HanMeimeiSay:
		log.Infof(m.Data)

		resp, _ := s.RequestWait(sourceId, &interval.LileiSay{
			Data: "Be my grilfriend?",
		})
		// waiting....
		log.Infof(resp.(*interval.HanMeimeiSay).Data)

		s.Request(sourceId, &interval.LileiSay{
			Data: "it's ok! I will protect you.",
		}).Handle(func(resp interface{}, err error) {
			log.Infof(resp.(*interval.HanMeimeiSay).Data)
		})

		s.AddTimer(tools.UUID(), tools.NowTime()+int64(1*time.Second), func(dt int64) {
			s.Request(sourceId, &interval.LileiSay{
				Data: "please~",
			}).Handle(func(resp interface{}, err error) {
				log.Infof(resp.(*interval.HanMeimeiSay).Data)
			})
		}, -1)
	}
}
func (s *Student) OnHandleRequest(sourceId, targetId string, requestId string, msg interface{}) error {
	switch m := msg.(type) {
	case *interval.LileiSay:
		log.Infof(m.Data)

		s.Response(requestId, &interval.HanMeimeiSay{
			Data: "no!",
		})
	}
	return nil
}
