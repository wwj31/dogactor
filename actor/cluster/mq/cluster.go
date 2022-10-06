package mq

import (
	"fmt"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/tools"
)

func WithRemote(url string) actor.SystemOption {
	return func(system *actor.System) error {
		cluster := newCluster(url)
		clusterActor := actor.New("cluster_"+tools.XUID(), cluster, actor.SetLocalized(), actor.SetMailBoxSize(5000))
		if e := system.Add(clusterActor); e != nil {
			return fmt.Errorf("%w %v", actorerr.RegisterClusterErr, e)
		}
		system.SetCluster(clusterActor)
		return nil
	}
}

func newCluster(url string) *Cluster {
	c := &Cluster{}

	return c
}

type Cluster struct {
	actor.Base

	mqURL string
}
