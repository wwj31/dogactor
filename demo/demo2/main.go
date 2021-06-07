package main

import (
	"github.com/wwj31/godactor/actor"
	"github.com/wwj31/godactor/actor/cluster"
	"github.com/wwj31/godactor/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	ETCD_ADDR   = "127.0.0.1:2379"
	ETCD_PREFIX = "demo/"
)

func main() {
	log.Init(log.TAG_DEBUG_I, nil, "./_log", "demo", 1)

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	system1, _ := actor.NewSystem(
		actor.Addr("127.0.0.1:1000"),
		cluster.WithRemote(ETCD_ADDR, ETCD_PREFIX),
	)
	system1.Regist(actor.New("demoA1", &DemoA{}))

	system2, _ := actor.NewSystem(
		actor.Addr("127.0.0.1:2000"),
		cluster.WithRemote(ETCD_ADDR, ETCD_PREFIX),
	)
	system2.Regist(actor.New("demoA2", &DemoA{}))
	system2.Regist(actor.New("demoB", &DemoB{}))

	select {
	case <-exit:
		system1.Stop()
		system2.Stop()
	}
	log.Stop()
}

type DemoA struct {
	actor.ActorHanlerBase
}

func (s *DemoA) Init() {
	s.AddTimer(time.Second*5, -1, func(dt int64) {
		if s.GetID() == "demoA1" {
			s.Request("demoA2", &DemoMsg{ModelName: s.GetID()}).Handle(func(resp interface{}, err error) {
				if err != nil {
					log.KV("error", err).Info("response from")
				} else {
					rep := resp.(*DemoMsg)
					log.KV("model", rep.ModelName).Info("response from")
				}

			})
		}
	})
}

func (s *DemoA) HandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	switch req := msg.(type) {
	case string:
		log.KV("msg", req).Info("recv req")
		time.Sleep(5 * time.Second)
		s.Response(requestId, "sleep 10")
	case *DemoMsg:
		log.KV("model", req.ModelName).Info("request from")
		s.Response(requestId, &DemoMsg{ModelName: s.GetID()})
	}
	return nil
}

type DemoB struct {
	actor.ActorHanlerBase
}

func (s *DemoB) Init() {
	log.Info("Init DemoB")
	s.AddTimer(time.Second*5, -1, func(dt int64) {
		resp, err := s.RequestWait("demoA1", &DemoMsg{ModelName: s.GetID()})
		log.KVs(log.Fields{"resp": resp, "err": err}).Info("demoB recv1")

		resp, err = s.RequestWait("demoA1", &DemoMsg{ModelName: s.GetID()})
		log.KVs(log.Fields{"resp": resp, "err": err}).Info("demoB recv2")
	})
}
