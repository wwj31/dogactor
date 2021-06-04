package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwj31/godactor/actor"
	"github.com/wwj31/godactor/actor/cluster"
	"github.com/wwj31/godactor/actor/cluster/remote_provider/remote_planc"
	"github.com/wwj31/godactor/actor/cluster/servmesh_provider/etcd"
	"github.com/wwj31/godactor/actor/cmd"
	"github.com/wwj31/godactor/actor/event"
	"github.com/wwj31/godactor/log"
)

func main() {
	log.Init(log.TAG_DEBUG_I, nil, "./_log", "demo", 1)

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	actorSystem1, _ := actor.Go(
		actor.Addr("127.0.0.1:1111"),
		actor.WithCMD(cmd.New()),
		actor.WithEvent(event.NewActorEvent()),
		actor.WithCluster(cluster.NewCluster(etcd.NewEtcd("127.0.0.1:2379", "demo/"), remote_planc.NewRemoteMgr())),
	)
	actorSystem1.Start()
	actorSystem1.Regist(actor.NewActor("demo1", &DemoA{}))

	actorSystem2, _ := actor.Go(
		actor.Addr("127.0.0.1:2222"),
		actor.WithCMD(cmd.New()),
		actor.WithEvent(event.NewActorEvent()),
		actor.WithCluster(cluster.NewCluster(etcd.NewEtcd("127.0.0.1:2379", "demo/"), remote_planc.NewRemoteMgr())),
	)
	actorSystem2.Start()
	actorSystem2.Regist(actor.NewActor("demo2", &DemoA{}))
	actorSystem2.Regist(actor.NewActor("demoB", &DemoB{}))
	select {
	case <-exit:
		actorSystem1.Stop()
		actorSystem2.Stop()
	}
	log.Stop()
}

type DemoA struct {
	actor.ActorHanlerBase
}

func (s *DemoA) Init() error {
	s.AddTimer(time.Second*5, -1, func(dt int64) {
		if s.GetID() == "demo1" {
			s.Request("demo2", &DemoMsg{ModelName: s.GetID()}).Handle(func(resp interface{}, err error) {
				if err != nil {
					log.KV("error", err).Info("response from")
				} else {
					rep := resp.(*DemoMsg)
					log.KV("model", rep.ModelName).Info("response from")
				}

			})
		}
	})
	return nil
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

func (s *DemoB) Init() error {
	log.Info("Init DemoB")
	s.AddTimer(time.Second*5, -1, func(dt int64) {
		resp, err := s.RequestWait("demo1", &DemoMsg{ModelName: s.GetID()})
		log.KVs(log.Fields{"resp": resp, "err": err}).Info("demoB recv1")

		resp, err = s.RequestWait("demo1111", &DemoMsg{ModelName: s.GetID()})
		log.KVs(log.Fields{"resp": resp, "err": err}).Info("demoB recv2")
	})
	return nil
}
