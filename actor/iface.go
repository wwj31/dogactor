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

		Timer
		Sender

		CallLua(name string, ret int, args ...lua.LValue) []lua.LValue

		RegistryCmd(cmd string, fn func(...string), usage ...string)
	}

	Timer interface {
		AddTimer(timeId string, endAt time.Time, callback func(dt time.Duration), triggerCount ...int) string
		CancelTimer(timerId string)
	}

	Sender interface {
		Send(targetId string, msg interface{}) error
		Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request)
		RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (result interface{}, err error)
		Response(requestId string, msg interface{}) error
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

		// OnHandleEvent process event
		OnHandleEvent(event interface{})
		// OnHandleMessage process message
		OnHandleMessage(sourceId, targetId string, msg interface{})
		// OnHandleRequest process request
		OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) error
	}

	// only Base implement
	actorInitiator interface {
		initActor(actor Actor)
	}
)
