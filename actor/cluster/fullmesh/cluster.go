package fullmesh

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/wwj31/dogactor/actor/event"
	"github.com/wwj31/dogactor/actor/internal"
	"github.com/wwj31/dogactor/actor/internal/innermsg"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote/conntcp"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/servmesh/etcd"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

func WithRemote(etcdAddr, prefix string) actor.SystemOption {
	return func(system *actor.System) error {
		addr := make(chan string, 1)
		cluster := newCluster(etcd.NewEtcd(etcdAddr, prefix), conntcp.NewRemoteMgr(), addr)
		clusterId := "cluster_" + tools.XUID()
		if e := system.NewActor(
			clusterId,
			cluster,
			actor.SetLocalized(),
			actor.SetMailBoxSize(2000),
		); e != nil {
			return fmt.Errorf("%w %v", actorerr.RegisterClusterErr, e)
		}
		go func() {
			system.Addr = <-addr
		}()
		system.SetCluster(clusterId)
		return nil
	}
}

func newCluster(cluster ServiceMeshProvider, remote RemoteProvider, randPort chan<- string) *Cluster {
	c := &Cluster{
		serviceMesh: cluster,
		remote:      remote,
		actors:      make(map[string]string),
		hosts:       make(map[string]map[string]bool),
		ready:       make(map[string]bool),
		RandAddr:    randPort,
	}

	return c
}

type Cluster struct {
	actor.Base
	RandAddr chan<- string
	canStop  bool

	serviceMesh ServiceMeshProvider
	remote      RemoteProvider

	actors map[string]string          //map[actorID]host
	hosts  map[string]map[string]bool //map[host]actorIds
	ready  map[string]bool            //map[host]true
}

func (c *Cluster) OnInit() {
	if err := c.remote.Start(c); err != nil {
		log.SysLog.Errorw("remote start error", "err", err)
	}

	if err := c.serviceMesh.Start(c); err != nil {
		log.SysLog.Errorw("serviceMesh start error", "err", err)
	}

	c.RandAddr <- c.remote.Addr()

	c.System().OnEvent(c.ID(), c.OnEventNewActor)
	c.System().OnEvent(c.ID(), c.OnEventDelActor)
	c.System().OnEvent(c.ID(), c.OnEventClusterUpdate)
	c.System().OnEvent(c.ID(), c.OnEventSessionClosed)
	c.System().OnEvent(c.ID(), c.OnEventSessionOpened)

	c.ready[c.remote.Addr()] = true
}

func (c *Cluster) OnStop() bool {
	return c.canStop
}
func (c *Cluster) stop() {
	c.canStop = true
	c.System().CancelAll(c.ID())
	c.serviceMesh.Stop()
	c.remote.Stop()
	c.Exit()
}

func (c *Cluster) OnHandle(msg actor.Message) {
	if c.ID() != msg.GetTargetId() {
		if err := c.sendRemote(msg.GetTargetId(), actor.RequestId(msg.GetRequestId()), msg.Payload().([]byte)); err != nil {
			log.SysLog.Errorw("remote actor send failed",
				"id", c.ID(),
				"sourceId", msg.GetSourceId(),
				"targetId", msg.GetTargetId(),
				"requestId", msg.GetRequestId(),
				"err", err,
			)
		}
		return
	}

	if str, ok := msg.Payload().(string); ok && str == "stop" {
		c.stop()
	} else {
		log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", str)
	}
	return
}

// OnNewServ dispatch a new remote node
func (c *Cluster) OnNewServ(actorId, host string, add bool) {
	c.System().DispatchEvent("", internal.EvClusterUpdate{ActorId: actorId, Host: host, Add: add})
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////

func (c *Cluster) OnSessionClosed(peerHost string) {
	c.System().DispatchEvent(c.ID(), internal.EvSessionClosed{PeerHost: peerHost})
}
func (c *Cluster) OnSessionOpened(peerHost string) {
	c.System().DispatchEvent(c.ID(), internal.EvSessionOpened{PeerHost: peerHost})
}

func (c *Cluster) OnSessionRecv(msg *innermsg.ActorMessage) {
	err := c.System().Send(msg.SourceId, msg.TargetId, actor.RequestId(msg.RequestId), msg)
	if err != nil {
		log.SysLog.Errorw("cluster OnSessionRecv send error", "msg", msg.String(), "err", err)
	}
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////

func (c *Cluster) sendRemote(targetId actor.Id, requestId actor.RequestId, bytes []byte) error {
	//Response的时候地址由requestId解析提供
	var addr string
	if reqSourceId, _, _addr, _ := requestId.Parse(); reqSourceId == targetId {
		addr = _addr
	} else if addr = c.actors[targetId]; addr == "" {
		// 本地没有，尝试去注册中心找一下
		var err error
		addr, err = c.serviceMesh.Get(targetId)
		if err != nil {
			return err
		}
	}
	return c.remote.SendMsg(addr, bytes)
}

func (c *Cluster) watchRemote(actorId actor.Id, host string, add bool) {
	if add {
		defer func() {
			log.SysLog.Infow("register remote actor",
				"host", host,
				"actorId", actorId,
				"ready", c.ready[host],
			)
			if c.ready[host] {
				c.System().DispatchEvent(c.ID(), event.EvNewActor{ActorId: actorId})
			}
		}()

		oldHost, ok := c.actors[actorId]
		if ok {
			if oldHost == host { //重复put
				return
			} else {
				c.delRemoteActor(actorId)
			}
		}

		c.actors[actorId] = host
		if host >= c.remote.Addr() {
			return
		}

		actors := c.hosts[host]
		if len(actors) > 0 { //已有client
			actors[actorId] = true
			return
		}

		c.hosts[host] = map[string]bool{actorId: true}
		c.remote.NewClient(host)
	} else {
		if oldHost := c.actors[actorId]; oldHost != host {
			return
		}
		c.delRemoteActor(actorId)
	}
}

func (c *Cluster) delRemoteActor(actorId actor.Id) {
	old := c.actors[actorId]
	delete(c.actors, actorId)

	c.System().DispatchEvent(c.ID(), event.EvDelActor{ActorId: actorId, FromCluster: true})

	actors, ok := c.hosts[old]
	if ok {
		delete(actors, actorId)
		if len(actors) == 0 {
			delete(c.hosts, old)
			delete(c.ready, old)
			c.remote.StopClient(old)
		}
	}

	log.SysLog.Infow("remove remote actor", "actor", actorId, "old host", old)
}

func (c *Cluster) OnEventNewActor(event internal.EvNewLocalActor) {
	if event.Publish {
		_ = c.serviceMesh.RegisterService(event.ActorId, c.remote.Addr())
		if event.Reg != nil {
			select {
			case event.Reg <- struct{}{}:
			default:
			}
		}
	}
}

func (c *Cluster) OnEventDelActor(ev event.EvDelActor) {
	if !ev.FromCluster && ev.Publish {
		_ = c.serviceMesh.UnregisterService(ev.ActorId)
	}

	// 等待requestWaiter退出后，才能退出Cluster
	if ev.ActorId == c.System().WaiterId() {
		log.SysLog.Infow("waiter was stopped")
		c.stop()
	}
}

func (c *Cluster) OnEventClusterUpdate(event internal.EvClusterUpdate) {
	c.watchRemote(event.ActorId, event.Host, event.Add)
}

func (c *Cluster) OnEventSessionClosed(ev internal.EvSessionClosed) {
	delete(c.ready, ev.PeerHost)
	for actorId, host := range c.actors {
		if host == ev.PeerHost {
			c.System().DispatchEvent(c.ID(), event.EvDelActor{ActorId: actorId, FromCluster: true})
		}
	}
}

func (c *Cluster) OnEventSessionOpened(ev internal.EvSessionOpened) {
	c.ready[ev.PeerHost] = true
	for actorId, host := range c.actors {
		if host == ev.PeerHost {
			c.System().DispatchEvent(c.ID(), event.EvNewActor{ActorId: actorId})
		}
	}
}

func (c *Cluster) clusterInfo() string {
	type Info struct {
		Host  string
		Actor string
	}

	var actors []Info
	for a, host := range c.actors {
		actors = append(actors, Info{
			Host:  host,
			Actor: a,
		})
	}
	sort.SliceStable(actors, func(i, j int) bool {
		return actors[i].Host < actors[j].Host
	})
	bytes, err := json.Marshal(actors)
	if err != nil {
		return fmt.Errorf("json marshal err:%v", err).Error()
	}
	return string(bytes)
}
