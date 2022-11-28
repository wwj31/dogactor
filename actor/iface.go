package actor

import (
	lua "github.com/yuin/gopher-lua"
	"time"
)

type (
	Actor interface {
		ID() string
		System() *System
		Exit()
		Drain(afterDrained ...func())

		Timer
		Sender

		CallLua(name string, ret int, args ...lua.LValue) []lua.LValue
	}

	Timer interface {
		AddTimer(timeId string, endAt time.Time, callback func(dt time.Duration), triggerCount ...int) string
		CancelTimer(timerId string)
	}

	Sender interface {
		Send(targetId Id, msg interface{}) error
		Request(targetId Id, msg interface{}, timeout ...time.Duration) (req *request)
		RequestWait(targetId Id, msg interface{}, timeout ...time.Duration) (result interface{}, err error)
		Response(requestId string, msg interface{}) error
	}

	Message interface {
		Free()
		GetSourceId() Id
		GetTargetId() Id
		GetRequestId() string
		GetMsgName() string
		Message() interface{}
		String() string
	}

	// spawnActor used to init actor
	spawnActor interface {
		actorHandler
		actorInitiator
	}

	// actorHandler
	actorHandler interface {
		OnInit()

		// OnStop true if right now stop，false if late stop
		OnStop() bool
		OnHandle(message Message)
	}

	// only Base implement
	actorInitiator interface {
		initActor(actor Actor)
	}
)
