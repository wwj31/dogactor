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
	Start(exceptPort ...int) error
	Port() int
	Stop()
}

type Client interface {
	SendMsg([]byte) error
	AddHandler(handler func() SessionHandler)
	Start(reconnect bool) error
	Stop()
}

type Session interface {
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

type SessionHandler interface {
	OnSessionCreated(Session)
	OnSessionClosed()
	OnRecv([]byte)
}

type DecodeEncoder interface {
	Decode([]byte) ([][]byte, error)
	Encode(data []byte) []byte
}

var sessionGenId uint64

func GenSessionId() uint64 {
	return atomic.AddUint64(&sessionGenId, 1)
}
