package cluster

import (
	"errors"
	"fmt"
	"github.com/wwj31/godactor/actor"
	"github.com/wwj31/godactor/actor/cluster/remote_provider/remote_planc"
	"github.com/wwj31/godactor/actor/cluster/servmesh_provider/etcd"
	"github.com/wwj31/godactor/actor/err"
	"reflect"
	"strings"

	"github.com/gogo/protobuf/proto"
)

func WithRemote(ectd_addr, prefix string) actor.SystemOption {
	return func(system *actor.ActorSystem) error {
		cluster := newCluster(etcd.NewEtcd(ectd_addr, prefix), remote_planc.NewRemoteMgr())
		actor := actor.New("cluster", cluster, actor.SetLocalized(), actor.SetMailBoxSize(5000))
		if e := system.Regist(actor); e != nil {
			return fmt.Errorf("%w %w", err.RegistClusterErr, e)
		}
		system.SetCluster(actor.GetID())
		return nil
	}
}

func newCluster(cluster IServiceMeshProvider, remote IRemoteProvider) *Cluster {
	c := &Cluster{
		serviceMesh: cluster,
		remote:      remote,
		actors:      make(map[string]string),
		clients:     make(map[string]map[string]bool),
		ready:       make(map[string]bool),
	}

	return c
}

type Cluster struct {
	actor.ActorHanlerBase

	serviceMesh IServiceMeshProvider
	remote      IRemoteProvider

	actors  map[string]string          //actorId=>host
	clients map[string]map[string]bool //host=>actorIds
	ready   map[string]bool            //host=>true
}

func (c *Cluster) Init() {
	c.ActorSystem().RegistEvent(c.GetID(), (*actor.Ev_newActor)(nil), (*actor.Ev_clusterUpdate)(nil), (*actor.Ev_newSession)(nil))

	if e := c.remote.Start(c.ActorSystem()); e != nil {
		logger.KV("err", e).Error("remote start error")
	}

	if e := c.serviceMesh.Start(c.ActorSystem()); e != nil {
		logger.KV("err", e).Error("serviceMesh start error")
	}

	c.RegistCmd("clusterinfo", c.clusterinfo)
	c.ready[c.ActorSystem().Address()] = true
}

func (c *Cluster) Stop() (immediatelyStop bool) {
	return false
}

func (c *Cluster) HandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	_, reqTargetId, _, _ := actor.ParseRequestId(requestId)
	if c.GetID() != reqTargetId {
		if err := c.sendRemote(sourceId, targetId, requestId, msg.(proto.Message)); err != nil {
			logger.KV("targetId", targetId).KV("error", err).Error("remote actor send failed")
			return err
		}
		return
	}
	return
}

func (c *Cluster) HandleMessage(sourceId, targetId string, msg interface{}) {
	if targetId != c.GetID() {
		if err := c.sendRemote(sourceId, targetId, "", msg.(proto.Message)); err != nil {
			logger.KV("targetId", targetId).KV("error", err).Error("remote actor send failed")
		}
	} else {
		switch message := msg.(type) {
		case string:
			if message == "stop" {
				c.serviceMesh.Stop()
				c.remote.Stop()
				c.LogicStop()
			}
		default:
			logger.KV("t", reflect.TypeOf(message).Name()).Warn("no such case type")
		}
	}
}

func (c *Cluster) sendRemote(sourceId, targetId, requestId string, actMsg proto.Message) error {
	//Response的时候地址由requestId解析提供
	var addr string
	if reqSourceId, _, _addr, _ := actor.ParseRequestId(requestId); reqSourceId == targetId {
		addr = _addr
	} else if addr = c.actors[targetId]; addr == "" {
		return errors.New("target actor not find")
	}
	return c.remote.Send(addr, sourceId, targetId, requestId, actMsg)

}

//单线程调用,否则考虑加锁
func (c *Cluster) watchRemote(actorId, host string, add bool) {
	if add {
		defer func() {
			logger.KV("host", host).KV("actorId", actorId).KV("ready", c.ready[host]).Debug("remote actor regist")
			if c.ready[host] {
				c.ActorSystem().DispatchEvent(c.GetID(), &actor.Ev_newActor{ActorId: actorId, FromCluster: true})
			}
		}()

		if old := c.actors[actorId]; old == host { //重复put
			return
		} else if _, ok := c.clients[old]; ok { //actor地址修改了
			c.delRemoteActor(actorId)
		}
		c.actors[actorId] = host
		if host >= c.ActorSystem().Address() {
			return
		}

		actors := c.clients[host]
		if len(actors) > 0 { //已有client
			actors[actorId] = true
			return
		}

		logger.KV("host", host).KV("actorId", actorId).Debug("try to connect")
		c.clients[host] = map[string]bool{actorId: true}
		c.remote.NewClient(host)
	} else {
		c.delRemoteActor(actorId)
	}
}

func (c *Cluster) delRemoteActor(actorId string) {
	old := c.actors[actorId]
	delete(c.actors, actorId)

	c.ActorSystem().DispatchEvent(c.GetID(), &actor.Ev_delActor{ActorId: actorId, FromCluster: true})

	if actors, ok := c.clients[old]; ok {
		delete(actors, actorId)
		if len(actors) == 0 {
			delete(c.clients, old)
			delete(c.ready, old)
			c.remote.StopClient(old)
		}
	}
}

func (c *Cluster) HandleEvent(event interface{}) {
	switch e := event.(type) {
	case *actor.Ev_newActor:
		if e.Publish {
			c.serviceMesh.RegistService(e.ActorId, c.ActorSystem().Address())
		}
	case *actor.Ev_clusterUpdate:
		c.watchRemote(e.ActorId, e.Host, e.Add)
	case *actor.Ev_newSession:
		c.ready[e.Host] = true
		logger.KV("host", e.Host).Debug("remote host connect")
		for actorId, host := range c.actors {
			if host == e.Host {
				c.ActorSystem().DispatchEvent(c.GetID(), &actor.Ev_newActor{ActorId: actorId, FromCluster: true})
			}
		}
	case *actor.Ev_delSession:
		delete(c.ready, e.Host)
		for actorId, host := range c.actors {
			if host == e.Host {
				c.ActorSystem().DispatchEvent(c.GetID(), &actor.Ev_delActor{ActorId: actorId, FromCluster: true})
			}
		}
	}
}

func (c *Cluster) clusterinfo(params ...string) {
	actors := []string{}
	for id, host := range c.actors {
		actors = append(actors, fmt.Sprintf("[actorId=%v host=%v]", id, host))
	}
	format := `
--------------------------------- remote actor ---------------------------------
 %s
--------------------------------- remote actor ---------------------------------
`
	logger.Info(fmt.Sprintf(format, strings.Join(actors, "\n")))
}
