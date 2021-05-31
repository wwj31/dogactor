package actor

import (
	"fmt"
	"github.com/wwj31/godactor/actor/err"
)

type (
	ClusterReq  struct{}
	ClusterResp struct {
		Actors []string
	}
)

func WithCluster(cluster IActorHandler) SystemOption {
	return func(system *ActorSystem) error {
		if system.event == nil {
			return err.ClusterNeedEventErr
		}
		actor := NewActor("cluster", cluster, SetLocalized(), SetMailBoxSize(5000))
		if e := system.Regist(actor); e != nil {
			return fmt.Errorf("%w %w", err.RegistClusterErr, e)
		}
		system.clusterInit.Add(1)
		system.clusterId = actor.GetID()
		return nil
	}
}
