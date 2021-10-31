package cluster

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/cluster/remote_provider/remote_grpc"
	"github.com/wwj31/dogactor/actor/cluster/servmesh_provider/etcd"
	"github.com/wwj31/dogactor/log"
)

func WithRemote(ectd_addr, prefix string) actor.SystemOption {
	return func(system *actor.System) error {
		cluster := newCluster(etcd.NewEtcd(ectd_addr, prefix), remote_grpc.NewRemoteMgr())
		clusterActor := actor.New("cluster", cluster, actor.SetLocalized(), actor.SetMailBoxSize(5000))
		if e := system.Regist(clusterActor); e != nil {
			return fmt.Errorf("%w %w", actorerr.RegistClusterErr, e)
		}
		system.SetCluster(clusterActor.GetID())
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
	actor.Base

	serviceMesh IServiceMeshProvider
	remote      IRemoteProvider

	actors  map[string]string          //actorId=>host
	clients map[string]map[string]bool //host=>actorIds
	ready   map[string]bool            //host=>true
}

func (c *Cluster) OnInit() {
	_ = c.System().RegistEvent(
		c.GetID(),
		(*actor.Ev_newActor)(nil),
		(*actor.Ev_clusterUpdate)(nil),
		(*actor.Ev_sessionClosed)(nil),
	)

	if e := c.remote.Start(c); e != nil {
		logger.KV("err", e).Error("remote start error")
	}

	if e := c.serviceMesh.Start(c); e != nil {
		logger.KV("err", e).Error("serviceMesh start error")
	}

	c.RegistCmd("clusterinfo", c.clusterinfo)
	c.ready[c.System().Address()] = true
}

func (c *Cluster) OnStop() (immediatelyStop bool) {
	return false
}

func (c *Cluster) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	_, reqTargetId, _, _ := actor.ParseRequestId(requestId)
	if c.GetID() != reqTargetId {
		if err := c.sendRemote(sourceId, targetId, requestId, msg.(proto.Message)); err != nil {
			logger.KVs(log.Fields{"id": c.GetID(), "targetId": targetId, "err": err}).Error("remote actor send failed")
			return err
		}
		return
	}
	return
}

func (c *Cluster) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	// cluster 只特殊处理 stop 消息，其余消息全部转发remote
	if targetId != c.GetID() {
		if e := c.sendRemote(sourceId, targetId, "", msg.(proto.Message)); e != nil {
			logger.KV("targetId", targetId).KV("error", e).Error("remote actor send failed")
		}
		return
	}

	if str, ok := msg.(string); ok && str == "stop" {
		c.serviceMesh.Stop()
		c.remote.Stop()
		c.Exit()
	} else {
		logger.KVs(log.Fields{"t": reflect.TypeOf(msg).Name(), "str": str}).Warn("no such case type")
	}
}

// 处理新服务
func (c *Cluster) OnNewServ(k, v string) {
	e := c.System().DispatchEvent("", &actor.Ev_clusterUpdate{ActorId: k, Host: v, Add: true})
	if e != nil {
		logger.KVs(log.Fields{"ActorId": k, "Host": v, "Add": true, "err": e}).Error("system dispatch event error")
	}
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////
func (c *Cluster) Address() string {
	return c.System().Address()
}
func (c *Cluster) OnSessionClosed(peerHost string) {
	_ = c.System().DispatchEvent(c.GetID(), &actor.Ev_sessionClosed{PeerHost: peerHost})
}
func (c *Cluster) OnSessionOpened(peerHost string) {
	_ = c.System().DispatchEvent(c.GetID(), &actor.Ev_sessionOpened{PeerHost: peerHost})
}
func (c *Cluster) OnSessionRecv(sourceId, targetId, requestId string, msg proto.Message) {
	e := c.System().Send(sourceId, targetId, requestId, msg)
	if e != nil {
		logger.KVs(log.Fields{"sourceId": sourceId, "targetId": targetId, "requestId": requestId, "err": e}).Debug("cluster OnSessionRecv send error")
	}
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////

func (c *Cluster) sendRemote(sourceId, targetId, requestId string, actMsg proto.Message) error {
	//Response的时候地址由requestId解析提供
	var addr string
	if reqSourceId, _, _addr, _ := actor.ParseRequestId(requestId); reqSourceId == targetId {
		addr = _addr
	} else if addr = c.actors[targetId]; addr == "" {
		return errors.New("target actor not find")
	}
	return c.remote.SendMsg(addr, sourceId, targetId, requestId, actMsg)

}

//单线程调用,否则考虑加锁
func (c *Cluster) watchRemote(actorId, host string, add bool) {
	if add {
		defer func() {
			logger.KV("host", host).KV("actorId", actorId).KV("ready", c.ready[host]).Debug("remote actor regist")
			if c.ready[host] {
				_ = c.System().DispatchEvent(c.GetID(), &actor.Ev_newActor{ActorId: actorId, FromCluster: true})
			}
		}()

		if old := c.actors[actorId]; old == host { //重复put
			return
		} else if _, ok := c.clients[old]; ok { //actor地址修改了
			c.delRemoteActor(actorId)
		}
		c.actors[actorId] = host
		if host >= c.System().Address() {
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

	_ = c.System().DispatchEvent(c.GetID(), &actor.Ev_delActor{ActorId: actorId, FromCluster: true})

	if actors, ok := c.clients[old]; ok {
		delete(actors, actorId)
		if len(actors) == 0 {
			delete(c.clients, old)
			delete(c.ready, old)
			c.remote.StopClient(old)
		}
	}
}

func (c *Cluster) OnHandleEvent(event interface{}) {
	switch e := event.(type) {
	case *actor.Ev_newActor:
		if e.Publish {
			_ = c.serviceMesh.RegistService(e.ActorId, c.System().Address())
		}
	case *actor.Ev_clusterUpdate:
		c.watchRemote(e.ActorId, e.Host, e.Add)
	case *actor.Ev_sessionClosed:
		delete(c.ready, e.PeerHost)
		for actorId, host := range c.actors {
			if host == e.PeerHost {
				_ = c.System().DispatchEvent(c.GetID(), &actor.Ev_delActor{ActorId: actorId, FromCluster: true})
			}
		}
	case *actor.Ev_sessionOpened:
		c.ready[e.PeerHost] = true
		logger.KV("host", e.PeerHost).Debug("remote host connect")
		for actorId, host := range c.actors {
			if host == e.PeerHost {
				_ = c.System().DispatchEvent(c.GetID(), &actor.Ev_newActor{ActorId: actorId, FromCluster: true})
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
