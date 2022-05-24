package network

import (
	"github.com/wwj31/dogactor/tools"
	"net"
	"sync/atomic"
)

type SessionType int

const (
	TypeTCP SessionType = 1
	TypeUDP SessionType = 2
	TypeWS  SessionType = 3
)

type Listener interface {
	Start() error
	Stop()
}

type Client interface {
	SendMsg([]byte) error
	AddLast(handler func() NetSessionHandler)
	Start(reconnect bool) error
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

var GenNetSessionId = genNetSessionId()

func genNetSessionId() func() uint32 {
	now := tools.Now()
	sessionGenId := uint32(now.Hour()*100000000 + now.Minute()*1000000 + now.Second()*10000)
	return func() uint32 { return atomic.AddUint32(&sessionGenId, 1) }
}
