package cluster

import (
	"errors"
	"fmt"
	"github.com/wwj31/dogactor/actor/cluster/remote_provider/remote_tcp"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"reflect"
	"sort"
	"strings"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/cluster/servmesh_provider/etcd"
)

func WithRemote(ectdAddr, prefix string) actor.SystemOption {
	return func(system *actor.System) error {
		cluster := newCluster(etcd.NewEtcd(ectdAddr, prefix), remote_tcp.NewRemoteMgr())
		clusterActor := actor.New("cluster_"+tools.XUID(), cluster, actor.SetLocalized(), actor.SetMailBoxSize(5000))
		if e := system.Add(clusterActor); e != nil {
			return fmt.Errorf("%w %v", actorerr.RegistClusterErr, e)
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
		clients:     make(map[string]map[string]bool),
		ready:       make(map[string]bool),
	}

	return c
}

type Cluster struct {
	actor.Base

	serviceMesh ServiceMeshProvider
	remote      RemoteProvider

	actors  map[string]string          //actorId=>host
	clients map[string]map[string]bool //host=>actorIds
	ready   map[string]bool            //host=>true
}

func (c *Cluster) OnInit() {
	_ = c.System().RegistEvent(
		c.ID(),
		actor.EvNewActor{},
		actor.EvDelActor{},
		actor.EvClusterUpdate{},
		actor.EvSessionClosed{},
	)

	if err := c.remote.Start(c); err != nil {
		log.SysLog.Errorw("remote start error", "err", err)
	}

	if err := c.serviceMesh.Start(c); err != nil {
		log.SysLog.Errorw("serviceMesh start error", "err", err)
	}

	c.RegistryCmd("clusterinfo", c.clusterinfo, "information of all cluster actor")
	c.ready[c.System().Address()] = true
}

func (c *Cluster) OnStop() bool {
	return false
}

func (c *Cluster) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	_, reqTargetId, _, _ := actor.ParseRequestId(requestId)
	if c.ID() != reqTargetId {
		if err := c.sendRemote(targetId, requestId, msg.([]byte)); err != nil {
			log.SysLog.Errorw("remote actor send failed",
				"id", c.ID(),
				"targetId", targetId,
				"err", err,
			)
			return err
		}
		return
	}
	return
}

func (c *Cluster) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	// cluster ??????????????? stop ?????????????????????????????????remote
	if targetId != c.ID() {
		if e := c.sendRemote(targetId, "", msg.([]byte)); e != nil {
			log.SysLog.Errorw("cluster handle message",
				"id", c.ID(),
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

//func (c *Cluster) OnSessionRecv(sourceId, targetId, requestId string, msg proto.Message) {
//	err := c.System().Send(sourceId, targetId, requestId, msg)
//	if err != nil {
//		log.SysLog.Errorw("cluster OnSessionRecv send error",
//			"sourceId", sourceId,
//			"targetId", targetId,
//			"requestId", requestId,
//			"err", err,
//		)
//	}
//}

func (c *Cluster) OnSessionRecv(msg *actor_msg.ActorMessage) {
	err := c.System().Send(msg.SourceId, msg.TargetId, msg.RequestId, msg)
	if err != nil {
		log.SysLog.Errorw("cluster OnSessionRecv send error", "msg", msg.String(), "err", err)
	}
}

////////////////////////////////////// RemoteHandler /////////////////////////////////////////////////////////////////

func (c *Cluster) sendRemote(targetId, requestId string, bytes []byte) error {
	//Response??????????????????requestId????????????
	var addr string
	if reqSourceId, _, _addr, _ := actor.ParseRequestId(requestId); reqSourceId == targetId {
		addr = _addr
	} else if addr = c.actors[targetId]; addr == "" {
		return errors.New("target actor not find")
	}
	return c.remote.SendMsg(addr, bytes)
}

func (c *Cluster) watchRemote(actorId, host string, add bool) {
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

		if old := c.actors[actorId]; old == host { //??????put
			return
		} else if _, ok := c.clients[old]; ok {
			c.delRemoteActor(actorId)
		}
		c.actors[actorId] = host
		if host >= c.System().Address() {
			return
		}

		actors := c.clients[host]
		if len(actors) > 0 { //??????client
			actors[actorId] = true
			return
		}

		c.clients[host] = map[string]bool{actorId: true}
		c.remote.NewClient(host)
	} else {
		c.delRemoteActor(actorId)
	}
}

func (c *Cluster) delRemoteActor(actorId string) {
	old := c.actors[actorId]
	delete(c.actors, actorId)

	c.System().DispatchEvent(c.ID(), actor.EvDelActor{ActorId: actorId, FromCluster: true})

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
	case actor.EvNewActor:
		if e.Publish {
			_ = c.serviceMesh.RegisterService(e.ActorId, c.System().Address())
		}
	case actor.EvDelActor:
		if !e.FromCluster && e.Publish {
			_ = c.serviceMesh.UnregisterService(e.ActorId)
		}
	case actor.EvClusterUpdate:
		c.watchRemote(e.ActorId, e.Host, e.Add)
	case actor.EvSessionClosed:
		delete(c.ready, e.PeerHost)
		for actorId, host := range c.actors {
			if host == e.PeerHost {
				c.System().DispatchEvent(c.ID(), actor.EvDelActor{ActorId: actorId, FromCluster: true})
			}
		}
	case actor.EvSessionOpened:
		c.ready[e.PeerHost] = true
		for actorId, host := range c.actors {
			if host == e.PeerHost {
				c.System().DispatchEvent(c.ID(), actor.EvNewActor{ActorId: actorId, FromCluster: true})
			}
		}
	}
}

func (c *Cluster) clusterinfo(params ...string) {
	actors := []string{}
	for id, host := range c.actors {
		actors = append(actors, fmt.Sprintf("???%-47v???%15v%2v", id, host, "???"))
	}
	sort.SliceStable(actors, func(i, j int) bool {
		return actors[i] < actors[j]
	})
	format := `
?????????????????????????????????????????????????????????????????????????????????remote actor?????????????????????????????????????????????????????????????????????????????????
???                   actorId                     ???     host       ???
??????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
%s
?????????????????????????????????????????????????????????????????????????????????remote actor?????????????????????????????????????????????????????????????????????????????????
`
	fmt.Println(fmt.Sprintf(format, strings.Join(actors, "\n")))
}
