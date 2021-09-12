package cluster

import (
	"github.com/golang/protobuf/proto"
	"github.com/wwj31/dogactor/actor/cluster/remote_provider"
	"github.com/wwj31/dogactor/actor/cluster/servmesh_provider"
)

type IServiceMeshProvider interface {
	Start(servmesh_provider.ServMeshHander) error
	Stop()
	RegistService(key string, value string) error
}

type IRemoteProvider interface {
	Start(remote_provider.RemoteHandler) error
	Stop()
	NewClient(host string)
	StopClient(host string)

	SendMsg(addr string, sourceId, targetId, requestId string, msg proto.Message) error
}
