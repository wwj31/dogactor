package fullmesh

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote_provider/remote_tcp"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/servmesh_provider/etcd"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"reflect"
	"sort"
)

func WithRemote(ectdAddr, prefix string) actor.SystemOption {
	return func(system *actor.System) error {
		cluster := newCluster(etcd.NewEtcd(ectdAddr, prefix), remote_tcp.NewRemoteMgr())
		clusterActor := actor.New("cluster_"+tools.XUID(), cluster, actor.SetLocalized(), actor.SetMailBoxSize(5000))
		if e := system.Add(clusterActor); e != nil {
			return fmt.Errorf("%w %v", actorerr.RegisterClusterErr, e)
		}
		system.SetCluster(clusterActor)
		return nil
	}
}

func newCluster(cluster ServiceMeshProvider, remote RemoteProvider) *Cluster {
	c := &Cluster{
		serviceMesh: cluster,
		remote:      remote,
		actors:      make(map[string]string),
		hosts:       make(map[string]map[string]bool),
		ready:       make(map[string]bool),
	}

	return c
}

type Cluster struct {
	actor.Base

	serviceMesh ServiceMeshProvider
	remote      RemoteProvider

	actors map[string]string          //actorId=>host
	hosts  map[string]map[string]bool //host=>actorIds
	ready  map[string]bool            //host=>true
}

func (c *Cluster) OnInit() {
	if err := c.remote.Start(c); err != nil {
		log.SysLog.Errorw("remote start error", "err", err)
	}

	if err := c.serviceMesh.Start(c); err != nil {
		log.SysLog.Errorw("serviceMesh start error", "err", err)
	}

	c.System().OnEvent(c.ID(), c.OnEventNewActor)
	c.System().OnEvent(c.ID(), c.OnEventDelActor)
	c.System().OnEvent(c.ID(), c.OnEventClusterUpdate)
	c.System().OnEvent(c.ID(), c.OnEventSessionClosed)
	c.System().OnEvent(c.ID(), c.OnEventSessionOpened)

	//c.RegistryCmd("clusterinfo", c.clusterinfo, "information of all cluster actor")
	c.ready[c.System().Address()] = true
}

func (c *Cluster) OnStop() bool {
	return false
}

func (c *Cluster) OnHandleRequest(sourceId, targetId actor.Id, requestId string, msg interface{}) (respErr error) {
	_, reqTargetId, _, _ := actor.ParseRequestId(requestId)
	if c.ID() != reqTargetId {
		if err := c.sendRemote(targetId, requestId, msg.([]byte)); err != nil {
			log.SysLog.Errorw("remote actor send failed",
				"id", c.ID(),
				"sourceId", sourceId,
				"targetId", targetId,
				"requestId", requestId,
				"err", err,
			)
			return err
		}
		return
	}

	str, ok := msg.(string)
	if ok {
		switch str {
		case "stop":
			c.System().CancelAll(c.ID())
			c.serviceMesh.Stop()
			c.remote.Stop()
			c.Exit()

		case "nodeinfo":
			respErr = c.Response(requestId, c.clusterInfo())
		default:
			log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", str)
		}
	}
	return
}

func (c *Cluster) OnHandleMessage(sourceId, targetId actor.Id, msg interface{}) {
	// cluster 只特殊处理 stop 消息，其余消息全部转发remote
	if targetId != c.ID() {
		if e := c.sendRemote(targetId, "", msg.([]byte)); e != nil {
			log.SysLog.Errorw("cluster handle message",
				"id", c.ID(),
				"sourceId", sourceId,
				"targetId", targetId,
				"err", e,
			)
		}
		return
	}

	if str, ok := msg.(string); ok && str == "stop" {
		c.System().CancelAll(c.ID())
		c.serviceMesh.Stop()
		c.remote.Stop()
		c.Exit()
	} else {
		log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", str)
	}
}

// OnNewServ dispatch a new remote
func (c *Cluster) OnNewServ(actorId, host string, add bool) {
	c.System().DispatchEvent("", actor.EvClusterUpdate{ActorId: actorId, Host: host, Add: add})
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////

func (c *Cluster) Address() string {
	return c.System().Address()
}
func (c *Cluster) OnSessionClosed(peerHost string) {
	c.System().DispatchEvent(c.ID(), actor.EvSessionClosed{PeerHost: peerHost})
}
func (c *Cluster) OnSessionOpened(peerHost string) {
	c.System().DispatchEvent(c.ID(), actor.EvSessionOpened{PeerHost: peerHost})
}

func (c *Cluster) OnSessionRecv(msg *actor_msg.ActorMessage) {
	err := c.System().Send(msg.SourceId, msg.TargetId, msg.RequestId, msg)
	if err != nil {
		log.SysLog.Errorw("cluster OnSessionRecv send error", "msg", msg.String(), "err", err)
	}
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////

func (c *Cluster) sendRemote(targetId actor.Id, requestId string, bytes []byte) error {
	//Response的时候地址由requestId解析提供
	var addr string
	if reqSourceId, _, _addr, _ := actor.ParseRequestId(requestId); reqSourceId == targetId {
		addr = _addr
	} else if addr = c.actors[targetId]; addr == "" {
		return errors.New("target actor not find")
	}
	return c.remote.SendMsg(addr, bytes)
}

func (c *Cluster) watchRemote(actorId actor.Id, host string, add bool) {
	if add {
		defer func() {
			log.SysLog.Infow("remote actor regist",
				"host", host,
				"actorId", actorId,
				"ready", c.ready[host],
			)
			if c.ready[host] {
				c.System().DispatchEvent(c.ID(), actor.EvNewActor{ActorId: actorId, FromCluster: true})
			}
		}()

		if old := c.actors[actorId]; old == host { //重复put
			return
		} else if _, ok := c.hosts[old]; ok {
			c.delRemoteActor(actorId)
		}
		c.actors[actorId] = host
		if host >= c.System().Address() {
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
		c.delRemoteActor(actorId)
	}
}

func (c *Cluster) delRemoteActor(actorId actor.Id) {
	old := c.actors[actorId]
	delete(c.actors, actorId)

	c.System().DispatchEvent(c.ID(), actor.EvDelActor{ActorId: actorId, FromCluster: true})

	if actors, ok := c.hosts[old]; ok {
		delete(actors, actorId)
		if len(actors) == 0 {
			delete(c.hosts, old)
			delete(c.ready, old)
			c.remote.StopClient(old)
		}
	}
}

func (c *Cluster) OnEventNewActor(event actor.EvNewActor) {
	if event.Publish {
		_ = c.serviceMesh.RegisterService(event.ActorId, c.System().Address())
	}
}

func (c *Cluster) OnEventDelActor(event actor.EvDelActor) {
	if !event.FromCluster && event.Publish {
		_ = c.serviceMesh.UnregisterService(event.ActorId)
	}
}

func (c *Cluster) OnEventClusterUpdate(event actor.EvClusterUpdate) {
	c.watchRemote(event.ActorId, event.Host, event.Add)
}

func (c *Cluster) OnEventSessionClosed(event actor.EvSessionClosed) {
	delete(c.ready, event.PeerHost)
	for actorId, host := range c.actors {
		if host == event.PeerHost {
			c.System().DispatchEvent(c.ID(), actor.EvDelActor{ActorId: actorId, FromCluster: true})
		}
	}
}

func (c *Cluster) OnEventSessionOpened(event actor.EvSessionOpened) {
	c.ready[event.PeerHost] = true
	for actorId, host := range c.actors {
		if host == event.PeerHost {
			c.System().DispatchEvent(c.ID(), actor.EvNewActor{ActorId: actorId, FromCluster: true})
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
