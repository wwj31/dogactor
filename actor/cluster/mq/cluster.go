package mq

import (
	"fmt"
	"reflect"

	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

func WithRemote(url string, mq MQ) actor.SystemOption {
	return func(system *actor.System) error {
		cluster := newCluster(url, mq)
		clusterActor := actor.New("cluster_"+tools.XUID(), cluster, actor.SetLocalized(), actor.SetMailBoxSize(1000))
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
		log.SysLog.Errorw("nat connect failed!", "url", c.mqURL)
		return
	}

	_ = c.System().RegistEvent(c.ID(),
		actor.EvNewActor{},
		actor.EvDelActor{},
	)
}

func (c *Cluster) OnStop() bool {
	return false
}

func (c *Cluster) stop() {
	c.System().CancelAll(c.ID())
	c.mq.Close()
	c.Exit()
}

func (c *Cluster) OnHandleRequest(sourceId, targetId actor.Id, requestId string, msg interface{}) (respErr error) {
	reqSourceId, reqTargetId, _, _ := actor.ParseRequestId(requestId)
	if c.ID() != reqTargetId {
		if sourceId == reqTargetId {
			targetId = reqSourceId
		} else {
			targetId = reqTargetId
		}

		err := c.mq.Pub(subFormat(targetId), msg.([]byte))
		if err != nil {
			log.SysLog.Errorw("remote actor send failed",
				"id", c.ID(),
				"sourceId", sourceId,
				"targetId", targetId,
				"requestId", requestId,
				"err", err,
			)
		}
		return
	}

	switch v := msg.(type) {
	case actor.ReqMsgDrain:
		err := c.mq.UnSub(sourceId)
		_ = c.Response(requestId, actor.RespMsgDrain{Err: err})

	case string:
		if v == "stop" {
			c.stop()
		} else {
			log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", v)
		}
	}
	return
}
func (c *Cluster) OnHandleMessage(sourceId, targetId actor.Id, msg interface{}) {
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
		c.stop()
	} else {
		log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", str)
	}
}

func (c *Cluster) OnHandleEvent(event interface{}) {
	switch e := event.(type) {
	case actor.EvNewActor:
		if e.Publish {
			err := c.mq.SubASync(subFormat(e.ActorId), func(data []byte) {
				msg := actor_msg.NewActorMessage() // remote recv message
				if err := msg.Unmarshal(data); err != nil {
					log.SysLog.Errorf("unmarshal msg failed", "err", err)
					return
				}
				expect.Nil(c.System().Send(msg.SourceId, msg.TargetId, msg.RequestId, msg))
			})

			if err != nil {
				log.SysLog.Errorw("mq cluster SubAsync failed!", "err", err, "event", e)
			}
		}

	case actor.EvDelActor:
		if e.Publish {
			err := c.mq.UnSub(subFormat(e.ActorId))
			if err != nil {
				log.SysLog.Errorw("mq cluster UnSub failed!", "err", err, "event", e)
			}
		}
	}
}

func subFormat(str actor.Id) string {
	return "mq:actor:" + str
}
