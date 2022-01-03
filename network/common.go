package network

import (
	"github.com/wwj31/dogactor/tools"
	"net"
	"sync/atomic"
)

type SessionType int

const (
	TYPE_TCP SessionType = 1
	TYPE_UDP SessionType = 2
	TYPE_WS  SessionType = 3
)

type Listener interface {
	Start() error
	Stop()
}

type Client interface {
	SendMsg([]byte) error
	AddLast(hander func() NetSessionHandler)
	Start(reconect bool) error
	Stop()
}

type NetSession interface {
	Id() uint32
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	RemoteIP() string
	SendMsg([]byte) error
	Stop()
	StoreKV(interface{}, interface{})
	DeleteKV(interface{})
	Load(interface{}) (interface{}, bool)
	Type() SessionType
}

type NetSessionHandler interface {
	OnSessionCreated(NetSession)
	OnSessionClosed()
	OnRecv([]byte)
}

var GenNetSessionId = _gen_net_session_id()

func _gen_net_session_id() func() uint32 {
	now := tools.Now()
	_session_gen_id := uint32(now.Hour()*100000000 + now.Minute()*1000000 + now.Second()*10000)
	return func() uint32 { return atomic.AddUint32(&_session_gen_id, 1) }
}
