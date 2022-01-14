package cluster

import (
	"github.com/wwj31/dogactor/actor/cluster/remote_provider"
	"github.com/wwj31/dogactor/actor/cluster/servmesh_provider"
)

type IServiceMeshProvider interface {
	Start(servmesh_provider.ServMeshHander) error
	Stop()
	RegisterService(key string, value string) error
	UnregisterService(key string) error
}

type IRemoteProvider interface {
	Start(remote_provider.RemoteHandler) error
	Stop()
	NewClient(host string)
	StopClient(host string)

	SendMsg(addr string, bytes []byte) error
}
