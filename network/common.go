package network

import (
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
	AddHandler(handler func() NetSessionHandler)
	Start(reconnect bool) error
	Stop()
}

type NetSession interface {
	Id() uint64
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

var sessionGenId uint64

func GenNetSessionId() uint64 {
	return atomic.AddUint64(&sessionGenId, 1)
}
