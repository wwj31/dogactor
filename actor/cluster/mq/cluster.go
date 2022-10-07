package mq

import (
	"fmt"

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

func (s *Cluster) OnInit() {

}
