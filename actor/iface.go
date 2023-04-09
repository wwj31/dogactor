package actor

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

type (
	Actor interface {
		ID() string
		System() *System
		Exit()
		AppendHandler(fn func(message Message) bool)
		CallLua(name string, ret int, args ...lua.LValue) []lua.LValue
	}

	Timer interface {
		AddTimer(timeId string, endAt time.Time, callback func(dt time.Duration), times ...int) string
		CancelTimer(timerId string)
		clear()
	}

	Messenger interface {
		Send(targetId Id, msg any) error
		Request(targetId Id, msg any, timeout ...time.Duration) (req *request)
		RequestWait(targetId Id, msg any, timeout ...time.Duration) (result interface{}, err error)
		Response(requestId string, msg any) error
	}

	Drainer interface {
		Drain(afterDrained ...func())
		isDraining() bool
	}

	Message interface {
		GetSourceId() Id
		GetTargetId() Id
		GetRequestId() string
		GetMsgName() string
		RawMsg() interface{}
		String() string
		Free()
	}

	// spawnActor used to init actor
	spawnActor interface {
		actorHandler
		actorInitiator
	}

	// actorHandler
	actorHandler interface {
		OnInit()
		OnStop() bool
		OnHandle(message Message)
	}

	// only Base implement
	actorInitiator interface {
		initActor(actor Actor)
	}
)
