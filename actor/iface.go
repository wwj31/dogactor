package actor

import (
	lua "github.com/yuin/gopher-lua"
	"time"
)

type (
	Actor interface {
		//core
		ID() string
		System() *System
		Exit()

		Timer
		Sender

		//lua
		CallLua(name string, ret int, args ...lua.LValue) []lua.LValue
		//cmd
		RegistCmd(cmd string, fn func(...string), usage ...string)
	}

	Timer interface {
		AddTimer(timeId string, endAt int64, callback func(dt int64), trigger_times ...int32) string
		UpdateTimer(timeId string, endAt int64) error
		CancelTimer(timerId string,del ...bool)
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
