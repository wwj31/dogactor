package actor

import (
	"math"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/internal/script"
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

type Id = string

type (
	Option func(*actor)
	actor  struct {
		system *System

		id      Id
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

		// record Request msg and delete after done
		requests map[string]*request

		// actor status schedule
		status    atomic.Value
		idleTimer *time.Timer

		// drain
		draining     atomic.Value
		afterDrained []func()
	}
)

// New build a new actor
// id is invalid if contain '@' or '$'
func New(id Id, handler spawnActor, opt ...Option) *actor {
	a := &actor{
		id:      id,
		handler: handler,
		mailBox: mailBox{
			ch: make(chan Message, 100),
		},
		remote:    true, // 默认都能被远端发现
		timerMgr:  jtimer.New(),
		timer:     globalTimerPool.Get(math.MaxInt),
		requests:  make(map[string]*request),
		idleTimer: globalTimerPool.Get(math.MaxInt),
	}
	a.status.Store(starting)
	a.draining.Store(false)
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
func (s *actor) Exit()           { _ = s.push(actorStop) }
func (s *actor) Drain(afterDrained ...func()) {
	s.Request(s.system.cluster.id, &ReqMsgDrain{}, 5*time.Minute).Handle(func(resp interface{}, err error) {
		if err != nil {
			log.SysLog.Errorw("drain failed ", "err", err)
			return
		}

		respMsgDrain, ok := resp.(*RespMsgDrain)
		if !ok {
			log.SysLog.Errorw("drain resp failed  ", "resp", reflect.TypeOf(resp).String())
			return
		}

		if respMsgDrain.Err != nil {
			log.SysLog.Errorw("drain return err ", "err", respMsgDrain.Err)
			return
		}

		s.afterDrained = afterDrained
		s.draining.Store(true)
	})
}

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

func (s *actor) push(msg Message) error {
	if msg == nil {
		return actorerr.ActorPushMsgErr
	}

	if l, c := len(s.mailBox.ch), cap(s.mailBox.ch); l > c*2/3 {
		log.SysLog.Warnw("mail box is almost full",
			"len", l,
			"cap", c,
			"actorId", s.id)
	}

	deadline := globalTimerPool.Get(time.Minute)
	select {
	case s.mailBox.ch <- msg:
	case <-deadline.C:
		log.SysLog.Warnw("mail box was full", "actorId", s.id)
	}
	globalTimerPool.Put(deadline)

	s.activate()
	return nil
}

var (
	timerTickMsg = &actor_msg.ActorMessage{MsgName: "timerTickMsg"} // timer tick msg
	actorStop    = &actor_msg.ActorMessage{MsgName: "actorStop"}    // actor stop msg
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
		log.SysLog.Infow("actor into running", "actor", s.ID())
		go s.run()
	}
}

func (s *actor) run() {
	s.resetIdleTime()
	for {
		select {
		case msg := <-s.mailBox.ch:
			if s.isStop(msg) {
				s.exit(true)
				return
			}

			tools.Try(func() {
				s.handleMsg(msg)

				// upset timer
				s.timerMgr.Update(tools.Now())
			})
			msg.Free()

			if s.draining.Load() == true && s.mailBox.Empty() {
				for _, fn := range s.afterDrained {
					fn()
				}

				s.exit(false)
				return
			}

			s.resetTime()
			s.resetIdleTime()
		case <-s.timer.C:
			_ = s.push(timerTickMsg)

		case <-s.idleTimer.C:
			if s.timerMgr.Len() > 0 {
				s.resetIdleTime()
				break
			}

			// check sent after the timeout
			if len(s.mailBox.ch) == 0 {
				s.status.Store(idle)
				log.SysLog.Infow("actor into idle", "actor", s.ID())
			}

			//double check
			if len(s.mailBox.ch) > 0 {
				s.activate()
			}

			// there are the return just for idle with the mailBox was empty and the timerMgr was empty
			return
		}
	}
}

func (s *actor) handleMsg(msg Message) {
	message := msg.Message()

	var msgType string
	// message is ActorMessage when msg from remote OnRecv
	if actMsg, ok := message.(*actor_msg.ActorMessage); ok {
		if actMsg.Data != nil && actMsg.MsgName != "" {
			defer actMsg.Free()
			msgType = actMsg.GetMsgName()
			message = actMsg.Fill(s.system.protoIndex)
			actMsg.SetMessage(message)
			msg = actMsg
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

	reqSourceId, _, _, _ := ParseRequestId(msg.GetRequestId())
	if s.id == reqSourceId {
		//handle Response
		s.doneRequest(msg.GetRequestId(), message)
		return
	}

	// fn
	if fn, ok := message.(func()); ok {
		fn()
		return
	}

	s.handler.OnHandle(msg)
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
		_ = s.push(actorStop)
	}
}

// resetIdleTime reset idleTimer
func (s *actor) resetIdleTime() {
	s.idleTimer = globalTimerPool.Get(time.Minute)
}

// resetTime reset timer of timerMgr
func (s *actor) resetTime() {
	nextAt := s.timerMgr.NextUpdateAt()
	if nextAt.Unix() == 0 {
		return
	}

	if !nextAt.Equal(s.nextAt) {
		s.timer = globalTimerPool.Get(nextAt.Sub(tools.Now()))
		s.nextAt = nextAt
	}
}

func (s *actor) isStop(msg Message) bool {
	message, ok := msg.(*actor_msg.ActorMessage)
	if ok && message == actorStop {
		return true
	}
	return false
}

func (s *actor) isDrained() bool {
	return s.draining.Load() == true && s.mailBox.Empty()
}

func (s *actor) exit(emitEvent bool) {
	log.SysLog.Infow("actor done", "actorId", s.ID())
	s.status.Store(stop)

	if emitEvent {
		s.system.DispatchEvent(s.id, EvDelActor{ActorId: s.id, Publish: s.remote})
	}

	s.system.actorCache.Delete(s.ID())
	s.system.waitStop.Done()
}

// Extra Option //////////

func SetMailBoxSize(boxSize int) Option {
	return func(a *actor) {
		a.mailBox.ch = make(chan Message, boxSize)
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
