package main

import (
	"github.com/wwj31/godactor/actor/cluster"
	"github.com/wwj31/godactor/demo/example2/interval"
	"github.com/wwj31/godactor/log"
	"net/http"
	"time"

	"github.com/wwj31/godactor/actor"
	_ "net/http/pprof"
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

func main() {
	log.Init(log.TAG_DEBUG_I, nil, "./_log", "demo", 1)
	system1, _ := actor.NewSystem(actor.Addr("127.0.0.1:1000"), cluster.WithRemote(ETCD_ADDR, ETCD_PREFIX))
	lilei := actor.New("LiLei", &Student{Name: "LiLei", Age: 19})
	system1.Regist(lilei)

	system2, _ := actor.NewSystem(actor.Addr("127.0.0.1:2000"), cluster.WithRemote(ETCD_ADDR, ETCD_PREFIX))
	hanmeimei := actor.New("HanMeimei", &Student{Name: "HanMeimei", Age: 15})
	system2.Regist(hanmeimei)

	err := http.ListenAndServe("0.0.0.0:6061", nil)
	if err != nil {
		log.Error(err.Error())
	}
	<-system1.CStop
	<-system2.CStop
}
func (s *Student) OnInit() {
	if s.Name == "LiLei" {
		s.AddTimer(2*time.Second, 1, func(dt int64) {
			s.Send("HanMeimei", &interval.LileiSay{Data: "hello, I'm Li Lei"})
		})
	}
}

func (s *Student) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch m := msg.(type) {
	case *interval.LileiSay:
		log.Info(m.Data)

		s.Send(sourceId, &interval.HanMeimeiSay{
			Data: "hi~! Li Lei, I'm HanMeimei",
		})
	case *interval.HanMeimeiSay:
		log.Info(m.Data)

		resp, _ := s.RequestWait(sourceId, &interval.LileiSay{
			Data: "Be my grilfriend?",
		})
		// waiting....
		log.Info(resp.(*interval.HanMeimeiSay).Data)

		s.Request(sourceId, &interval.LileiSay{
			Data: "it's ok! I will protect you.",
		}).Handle(func(resp interface{}, err error) {
			log.Info(resp.(*interval.HanMeimeiSay).Data)
		})

		s.AddTimer(time.Duration(10)*time.Millisecond, -1, func(dt int64) {
			s.Request(sourceId, &interval.LileiSay{
				Data: "please~",
			}).Handle(func(resp interface{}, err error) {
				log.Info(resp.(*interval.HanMeimeiSay).Data)
			})
		})
	}
}
func (s *Student) OnHandleRequest(sourceId, targetId string, requestId string, msg interface{}) error {
	switch m := msg.(type) {
	case *interval.LileiSay:
		log.Info(m.Data)

		s.Response(requestId, &interval.HanMeimeiSay{
			Data: "no!",
		})
	}
	return nil
}
