package mq

import (
	"fmt"
	"github.com/wwj31/dogactor/actor/event"
	"github.com/wwj31/dogactor/actor/internal"

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
		id := "cluster_" + tools.XUID()
		if e := system.NewActor(
			id,
			cluster,
			actor.SetLocalized(),
			actor.SetMailBoxSize(1000),
		); e != nil {
			return fmt.Errorf("%w %v", actorerr.RegisterClusterErr, e)
		}
		system.SetCluster(id)
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

	c.System().OnEvent(c.ID(), func(ev event.EvNewActor) {
		if ev.Publish {
			err := c.mq.SubASync(subFormat(ev.ActorId), func(data []byte) {
				msg := actor_msg.NewActorMessage() // remote recv message
				if err := msg.Unmarshal(data); err != nil {
					log.SysLog.Errorf("unmarshal msg failed", "err", err)
					return
				}
				expect.Nil(c.System().Send(msg.SourceId, msg.TargetId, actor.RequestId(msg.RequestId), msg))
			})

			if err != nil {
				log.SysLog.Errorw("mq cluster SubAsync failed!", "err", err, "event", ev)
			}
			c.System().DispatchEvent(c.ID(), internal.EvActorSubMqFin{ActorId: ev.ActorId})
		}
	})

	c.System().OnEvent(c.ID(), func(ev event.EvDelActor) {
		if ev.Publish {
			err := c.mq.UnSub(subFormat(ev.ActorId), false)
			if err != nil {
				log.SysLog.Errorw("mq cluster UnSub failed!", "err", err, "event", ev)
			}
		}

		if ev.ActorId == c.System().WaiterId() {
			log.SysLog.Infow("waiter was stopped")
			c.stop()
		}
	})
}

func (c *Cluster) OnHandle(msg actor.Message) {
	if msg.GetTargetId() != c.ID() {
		if e := c.mq.Pub(subFormat(msg.GetTargetId()), msg.RawMsg().([]byte)); e != nil {
			log.SysLog.Errorw("cluster handle message",
				"id", c.ID(),
				"sourceId", msg.GetSourceId(),
				"targetId", msg.GetTargetId(),
				"err", e,
			)
		}
		return
	}

	switch msg.RawMsg().(type) {
	case *internal.ReqMsgDrain:
		err := c.mq.UnSub(subFormat(msg.GetSourceId()), true)
		_ = c.Response(msg.GetRequestId(), &internal.RespMsgDrain{Err: err})
	}
}

func (c *Cluster) OnStop() bool {
	return false
}

func (c *Cluster) stop() {
	c.System().CancelAll(c.ID())
	c.mq.Close()
	c.Exit()
}

func subFormat(str actor.Id) string {
	return "mq:actor:" + str
}
