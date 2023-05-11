package fullmesh

import (
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/servmesh"
)

type ServiceMeshProvider interface {
	Start(servmesh.ServMeshHander) error
	Stop()
	RegisterService(key string, value string) error
	UnregisterService(key string) error
}

type RemoteProvider interface {
	Start(remote.Handler) error
	Addr() string
	Stop()
	NewClient(host string)
	StopClient(host string)

	SendMsg(addr string, bytes []byte) error
}
