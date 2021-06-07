package cluster

import (
	"github.com/golang/protobuf/proto"
	"github.com/wwj31/godactor/actor"
)

type IServiceMeshProvider interface {
	Start(actorSystem *actor.System) error
	Stop()
	RegistService(key string, value string) error
}

type IRemoteProvider interface {
	Start(actorSystem *actor.System) error
	Stop()
	NewClient(host string)
	StopClient(host string)

	Send(addr string, sourceId, targetId, requestId string, actMsg proto.Message) error
}
