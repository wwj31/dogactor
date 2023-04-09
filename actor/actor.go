package actor

import (
	"github.com/wwj31/dogactor/actor/event"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/internal/script"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/timer"
	"github.com/wwj31/dogactor/tools"
)

var stopMsg = &actor_msg.ActorMessage{MsgName: "actorStop"}

const (
	starting = iota + 1
	idle
	running
	stop
)

type Id = string

type actor struct {
	system *System

	id       Id
	handler  actorHandler
	msgChain []func(message Message) bool
	mailBox  mailBox

	remote bool

	//lua
	lua     script.ILua
	luaPath string

	// actor status schedule
	status         atomic.Value
	handleLatestAt time.Time
}

func (s *actor) ID() string                                  { return s.id }
func (s *actor) System() *System                             { return s.system }
func (s *actor) Exit()                                       { _ = s.push(stopMsg) }
func (s *actor) AppendHandler(fn func(message Message) bool) { s.msgChain = append(s.msgChain, fn) }

func (s *actor) push(msg Message) error {
	if msg == nil {
		return actorerr.ActorPushMsgErr
	}

	if l, c := len(s.mailBox.ch), cap(s.mailBox.ch); l > c*2/3 {
		log.SysLog.Warnw("mail box is almost full", "len", l, "cap", c, "actorId", s.id)
	}

	deadline := timer.Get(time.Minute)
	select {
	case s.mailBox.ch <- msg:
	case <-deadline.C:
		log.SysLog.Warnw("mail box was full", "actorId", s.id)
	}
	timer.Put(deadline)

	s.activate()
	return nil
}

func (s *actor) init(ok chan<- struct{}) {
	tools.Try(s.handler.OnInit)
	if ok != nil {
		ok <- struct{}{}
	}
	s.status.CompareAndSwap(starting, idle)
	s.system.DispatchEvent(s.id, event.EvNewActor{ActorId: s.id, Publish: s.remote})
	s.activate()
}

func (s *actor) activate() {
	if s.status.CompareAndSwap(idle, running) {
		log.SysLog.Infow("actor into running", "actor", s.ID())
		go s.run()
	}
}

// actor event loop
func (s *actor) run() {
	idleTicker := time.NewTicker(time.Minute)
	defer idleTicker.Stop()

	for {
		select {
		case msg := <-s.mailBox.ch:
			if s.isStop(msg) {
				s.exit()
				return
			}

			tools.Try(func() { s.handleMsg(msg) })
			s.handleLatestAt = tools.Now()
			msg.Free()

		case <-idleTicker.C:
			if len(s.mailBox.ch) == 0 && tools.Now().Sub(s.handleLatestAt) > 5*time.Minute {
				s.status.Store(idle)
				log.SysLog.Infow("actor into idle", "actor", s.ID())

				//double check
				if len(s.mailBox.ch) > 0 {
					s.activate()
				}

				// there are the return just for idle state when both the mailBox and the timerMgr are empty
				return
			}

		}
	}
}

func (s *actor) handleMsg(msg Message) {
	rawMsg := msg.RawMsg()
	if rawMsg == nil {
		return
	}

	// if the message comes from a cluster, it can be asserted
	// that it's an ActorMessage representing that the message
	// was sent from a remote location,otherwise,it can be used directly.
	if actMsg, ok := rawMsg.(*actor_msg.ActorMessage); ok {
		if actMsg.MsgName != "" {
			defer actMsg.Free()
			rawMsg = actMsg.Fill(s.system.protoIndex)
			actMsg.SetMessage(rawMsg)
			msg = actMsg
		}
	}

	defer s.mailBox.recording(tools.Now(), msg.GetMsgName())

	for i := len(s.msgChain) - 1; i >= 0; i-- {
		if !s.msgChain[i](msg) {
			break
		}
	}
}

func (s *actor) stop() {
	var canceled bool
	tools.Try(func() {
		canceled = s.handler.OnStop()
	}, func(ex interface{}) {
		canceled = true
	})

	if canceled {
		_ = s.push(stopMsg)
	}
}

func (s *actor) isStop(msg Message) bool {
	message, ok := msg.(*actor_msg.ActorMessage)
	if ok && message == stopMsg {
		return true
	}
	return false
}

func (s *actor) exit() {
	log.SysLog.Infow("actor done", "actorId", s.ID())
	s.status.Store(stop)

	s.system.DispatchEvent(s.id, event.EvDelActor{ActorId: s.id, Publish: s.remote})
	s.system.actorCache.Delete(s.ID())
	s.system.waitStop.Done()
}

// Option the extra options of the new actor
type Option func(*actor)

func SetMailBoxSize(boxSize int) Option {
	return func(a *actor) {
		a.mailBox.ch = make(chan Message, boxSize)
	}
}

// SetLocalized indicates that it can't be discovered
// by other System, meaning, this actor is localized.
func SetLocalized() Option {
	return func(a *actor) {
		a.remote = false
	}
}

func SetLua(path string) Option {
	return func(a *actor) {
		a.lua = script.New()
		a.register2Lua()
		a.luaPath = path
		a.lua.Load(a.luaPath)
	}
}
