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

		//Timer 计时器
		Timer

		// Sender 不保证消息发送可靠性
		Sender

		//lua
		CallLua(name string, ret int, args ...lua.LValue) []lua.LValue
		//cmd
		RegistCmd(cmd string, fn func(...string), usage ...string)
	}

	Timer interface {
		AddTimer(timeId string, interval time.Duration, callback func(dt int64), trigger_times ...int32) string
		CancelTimer(timerId string)
	}

	Sender interface {
		Send(targetId string, msg interface{}) error
		Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request)
		RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (result interface{}, err error)
		Response(requestId string, msg interface{}) error
	}

	// spawnActor 基于携带匿名 Base 的结构
	spawnActor interface {
		actorHandler
		actorInitiator
	}

	// actorHandler
	actorHandler interface {
		OnInit()

		// OnStop true 立刻停止，false 延迟停止
		OnStop() bool

		// OnHandleEvent 事件消息
		OnHandleEvent(event interface{})
		// OnHandleMessage 普通消息
		OnHandleMessage(sourceId, targetId string, msg interface{})
		// OnHandleRequest 请求消息(需要应答)
		OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) error
	}

	// 仅 Base 实现
	actorInitiator interface {
		initActor(actor Actor)
	}
)
