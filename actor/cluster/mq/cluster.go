package mq

import (
	"fmt"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"reflect"

	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/tools"
)

func WithRemote(url string, mq MQ) actor.SystemOption {
	return func(system *actor.System) error {
		cluster := newCluster(url, mq)
		clusterActor := actor.New("cluster_"+tools.XUID(), cluster, actor.SetLocalized(), actor.SetMailBoxSize(5000))
		if e := system.Add(clusterActor); e != nil {
			return fmt.Errorf("%w %v", actorerr.RegisterClusterErr, e)
		}
		system.SetCluster(clusterActor)
		return nil
	}
}

func newCluster(url string, mq MQ) *Cluster {
	c := &Cluster{
		mqURL: nats.DefaultURL,
		mq:    mq,
	}
	if url != "" {
		c.mqURL = url
	}

	return c
}

type Cluster struct {
	actor.Base

	mqURL string
	mq    MQ
}

func (c *Cluster) OnInit() {
	err := c.mq.Connect(c.mqURL)
	if err != nil {
		log.SysLog.Errorf("nat connect failed!", "url", c.mqURL)
		return
	}

	_ = c.System().RegistEvent(c.ID(),
		actor.EvNewActor{},
		actor.EvDelActor{},
	)
}

func (c *Cluster) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	// cluster 只特殊处理 stop 消息，其余消息全部转发remote
	if targetId != c.ID() {
		if e := c.mq.Pub(subFormat(targetId), msg.([]byte)); e != nil {
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
		c.mq.Close()
		c.Exit()
	} else {
		log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", str)
	}
}
func (c *Cluster) OnHandleEvent(event interface{}) {
	switch e := event.(type) {
	case actor.EvNewActor:
		if e.Publish {
			err := c.mq.SubASync(subFormat(e.ActorId), func(msg MSG) {
				expect.Nil(c.Send(e.ActorId, msg))
			})

			if err != nil {
				log.SysLog.Errorf("mq cluster SubAsync failed!", "err", err, "event", e)
			}
		}

	case actor.EvDelActor:
		if e.Publish {
			err := c.mq.UnSub(subFormat(e.ActorId))
			if err != nil {
				log.SysLog.Errorf("mq cluster UnSub failed!", "err", err, "event", e)
			}
		}
	}
}

func subFormat(str string) string {
	return "mq.actor." + str
}
