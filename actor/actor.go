package actor

import (
	"context"
	"math"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/internal/script"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"github.com/wwj31/jtimer"
)

const (
	starting = iota + 1
	idle
	running
	stop
)

const idleTimeout = time.Minute

type (
	Option func(*actor)
	actor  struct {
		system *System

		id      string
		handler actorHandler
		mailBox mailBox

		remote bool

		// timer
		timerMgr *jtimer.Manager
		timer    *time.Timer
		nextAt   time.Time

		//lua
		lua     script.ILua
		luapath string

		// record Request msg and del in done
		requests map[string]*request

		// actor status control
		status    atomic.Value
		idleTimer *time.Timer

		// span context
		ctx context.Context
	}
)

// New build a new actor
// id is invalid if contain '@' or '$'
func New(id string, handler spawnActor, opt ...Option) *actor {
	a := &actor{
		id:      id,
		handler: handler,
		mailBox: mailBox{
			ch: make(chan actor_msg.Message, 100),
		},
		remote:    true, // 默认都能被远端发现
		timerMgr:  jtimer.New(),
		timer:     time.NewTimer(math.MaxInt),
		requests:  make(map[string]*request),
		idleTimer: time.NewTimer(math.MaxInt),
	}
	a.status.Store(starting)
	a.timer.Stop()
	a.idleTimer.Stop()

	handler.initActor(a)

	for _, f := range opt {
		f(a)
	}
	return a
}

func (s *actor) ID() string      { return s.id }
func (s *actor) System() *System { return s.system }
func (s *actor) Exit()           { _ = s.push(actorStopMsg) }

func (s *actor) Send(targetId string, msg interface{}) error {
	return s.system.Send(s.id, targetId, "", msg)
}

// AddTimer default value of timeId is uuid.
// if count == -1 timer will repack in timer system infinitely util call CancelTimer,
// when timeId is existed the timer will update with endAt and callback and count.
func (s *actor) AddTimer(timeId string, endAt time.Time, callback func(dt time.Duration), count ...int) string {
	c := 1
	if len(count) > 0 {
		c = count[0]
	}

	timeId = s.timerMgr.Add(tools.Now(), endAt, callback, c, timeId)

	s.resetTime()
	s.activate()
	return timeId
}

// CancelTimer remote a timer with
func (s *actor) CancelTimer(timerId string) {
	s.timerMgr.Remove(timerId)
	s.resetTime()
}

func (s *actor) push(msg actor_msg.Message) error {
	if msg == nil {
		return actorerr.ActorPushMsgErr
	}

	if l, c := len(s.mailBox.ch), cap(s.mailBox.ch); l > c*2/3 {
		log.SysLog.Warnw("mail box will quickly full",
			"len", l,
			"cap", c,
			"actorId", s.id)
	}

	s.mailBox.ch <- msg
	s.activate()
	return nil
}

var (
	timerTickMsg = &actor_msg.ActorMessage{MsgName: "timerTickMsg"} // timer tick msg
	actorStopMsg = &actor_msg.ActorMessage{MsgName: "actorStopMsg"} // actor stop msg
)

func (s *actor) init(ok chan<- struct{}) {
	tools.Try(s.handler.OnInit)
	if ok != nil {
		ok <- struct{}{}
	}
	s.status.CompareAndSwap(starting, idle)
	s.system.DispatchEvent(s.id, EvNewActor{ActorId: s.id, Publish: s.remote})
	s.activate()
}

func (s *actor) activate() {
	if s.status.CompareAndSwap(idle, running) {
		go s.run()
	}
}

func (s *actor) run() {
	s.resetIdleTime()
	for {
		select {
		case msg := <-s.mailBox.ch:
			if s.isStop(msg) {
				s.exit()
				return
			}

			tools.Try(func() {
				s.handleMsg(msg)

				// upset timer
				s.timerMgr.Update(tools.Now())
			})

			msg.Free()
			s.resetTime()
			s.resetIdleTime()

		case <-s.timer.C:
			_ = s.push(timerTickMsg)

		case <-s.idleTimer.C:
			if s.timerMgr.Len() > 0 {
				s.resetIdleTime()
				break
			}
			s.status.Store(idle)

			// check sent after the timeout
			if len(s.mailBox.ch) > 0 {
				s.activate()
			}
			// there are the return just for idle with the mailBox was empty and the timerMgr was empty
			return
		}
	}
}

func (s *actor) handleMsg(msg actor_msg.Message) {
	message := msg.Message()

	var msgType string
	// message is ActorMessage when msg from remote OnRecv
	if actMsg, ok := message.(*actor_msg.ActorMessage); ok {
		if actMsg.Data != nil && actMsg.MsgName != "" {
			defer msg.Free()

			msgType = actMsg.GetMsgName()
			message = actMsg.Fill(s.system.protoIndex)
		}
	}
	if message == nil {
		return
	}

	if msgType == "" {
		msgType = reflect.TypeOf(message).String()
	}

	s.mailBox.processingTime = 0
	defer s.mailBox.recording(time.Now(), msgType)

	reqSourceId, _, _, reqOK := ParseRequestId(msg.GetRequestId())
	if reqOK {
		//rev Response
		if s.id == reqSourceId {
			s.doneRequest(msg.GetRequestId(), message)
			return
		}
		// rev Request
		if e := s.handler.OnHandleRequest(msg.GetSourceId(), msg.GetTargetId(), msg.GetRequestId(), message); e != nil {
			expect.Nil(s.Response(msg.GetRequestId(), &actor_msg.RequestDeadLetter{Err: e.Error()}))
		}
		return
	}

	//message
	if event, ok := message.(*actor_msg.EventMessage); ok {
		s.handler.OnHandleEvent(event.ActEvent())
		return
	}

	// fn
	if fn, ok := message.(func()); ok {
		fn()
		return
	}

	s.handler.OnHandleMessage(msg.GetSourceId(), msg.GetTargetId(), message)
}

func (s *actor) isWaitActor() bool {
	_, ok := s.handler.(*waitActor)
	return ok
}

// system close
func (s *actor) stop() {
	var stop bool
	tools.Try(func() {
		stop = s.handler.OnStop()
	}, func(ex interface{}) {
		stop = true
	})

	if stop {
		_ = s.push(actorStopMsg)
	}
}

// resetIdleTime reset idleTimer
func (s *actor) resetIdleTime() {
	if !s.idleTimer.Stop() {
		select {
		case <-s.idleTimer.C:
		default:
		}
	}
	s.idleTimer.Reset(idleTimeout)
}

// resetTime reset timer of timerMgr
func (s *actor) resetTime() {
	nextAt := s.timerMgr.NextUpdateAt()
	if nextAt.Unix() == 0 {
		return
	}

	if !nextAt.Equal(s.nextAt) {
		if !s.timer.Stop() {
			select {
			case <-s.timer.C:
			default:
			}
		}
		s.timer.Reset(nextAt.Sub(tools.Now()))
		s.nextAt = nextAt
	}
}

func (s *actor) isStop(msg actor_msg.Message) bool {
	message, ok := msg.(*actor_msg.ActorMessage)
	if ok && message == actorStopMsg {
		return true
	}
	return false
}

func (s *actor) exit() {
	log.SysLog.Infow("actor done", "actorId", s.ID())
	s.status.Store(stop)

	s.system.DispatchEvent(s.id, EvDelActor{ActorId: s.id, Publish: s.remote})
	s.system.actorCache.Delete(s.ID())
	s.system.waitStop.Done()
}

// Extra Option //////////

func SetMailBoxSize(boxSize int) Option {
	return func(a *actor) {
		a.mailBox.ch = make(chan actor_msg.Message, boxSize)
	}
}

func SetLocalized() Option {
	return func(a *actor) {
		a.remote = false
	}
}

func SetLua(path string) Option {
	return func(a *actor) {
		a.lua = script.New()
		a.register2Lua()
		a.luapath = path
		a.lua.Load(a.luapath)
	}
}
