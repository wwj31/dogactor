package actor

import (
	"github.com/wwj31/jtimer"
	"math"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/internal/script"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

const (
	idle    = 1
	running = 2
	stop    = 3
)

const idleTimeout = time.Minute

type (
	Option func(*actor)
	actor  struct {
		system *System

		id      string
		handler actorHandler
		mailBox chan actor_msg.Message
		remote  bool

		// timer
		timerMgr jtimer.TimerMgr
		timer    *time.Timer
		nextAt   int64

		//lua
		lua     script.ILua
		luapath string

		// record Request msg and del in done
		requests map[string]*request

		// actor status control
		status    atomic.Value
		idleTimer *time.Timer
	}
)

// New new actor
// id is invalid if contain '@' or '$'
func New(id string, handler spawnActor, op ...Option) *actor {
	a := &actor{
		id:        id,
		handler:   handler,
		mailBox:   make(chan actor_msg.Message, 100),
		remote:    true, // 默认都能被远端发现
		timerMgr:  jtimer.NewTimerMgr(),
		timer:     time.NewTimer(math.MaxInt),
		requests:  make(map[string]*request),
		idleTimer: time.NewTimer(math.MaxInt),
	}
	a.status.Store(idle)
	a.timer.Stop()
	a.idleTimer.Stop()

	handler.initActor(a)

	for _, f := range op {
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
func (s *actor) AddTimer(timeId string, endAt int64, callback func(dt int64), count ...int32) string {
	now := tools.NowTime()
	c := int32(1)
	if len(count) > 0 {
		c = count[0]
	}
	var (
		err      error
		newTimer *jtimer.Timer
	)

	newTimer, err = jtimer.NewTimer(now, endAt, c, callback, timeId)
	if err != nil {
		log.SysLog.Errorw("AddTimer new timer failed", "err", err)
		return ""
	}

	timeId, err = s.timerMgr.AddTimer(newTimer)
	if err != nil {
		log.SysLog.Errorw("AddTimer add failed", "err", err)
		return ""
	}
	s.resetTime()
	s.activate()
	return timeId
}

func (s *actor) UpdateTimer(timeId string, endAt int64) error {
	err := s.timerMgr.UpdateTimer(timeId, endAt)
	if err != nil {
		return err
	}
	s.resetTime()
	return nil
}

// 删除一个定时器
func (s *actor) CancelTimer(timerId string, del ...bool) {
	s.timerMgr.CancelTimer(timerId, del...)
	s.resetTime()
}

// Push一个消息
func (s *actor) push(msg actor_msg.Message) error {
	if msg == nil {
		return actorerr.ActorPushMsgErr
	}

	if l, c := len(s.mailBox), cap(s.mailBox); l > c*2/3 {
		log.SysLog.Warnw("mail box will quickly full",
			"len", l,
			"cap", c,
			"actorId", s.id)
	}

	s.mailBox <- msg
	s.activate()
	return nil
}

var (
	timerTickMsg = &actor_msg.ActorMessage{} // timer tick msg
	actorStopMsg = &actor_msg.ActorMessage{} // actor stop msg
)

func (s *actor) init(ok chan<- struct{}) {
	tools.Try(s.handler.OnInit)
	if ok != nil {
		ok <- struct{}{}
	}

	s.system.DispatchEvent(s.id, EvNewactor{ActorId: s.id, Publish: s.remote})
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
		case msg := <-s.mailBox:
			if s.isStop(msg) {
				s.exit()
				return
			}

			tools.Try(func() {
				s.handleMsg(msg)

				// upset timer
				s.timerMgr.Update(tools.NowTime())
			})

			msg.Free()
			s.resetTime()
			s.resetIdleTime()

		case <-s.timer.C:
			_ = s.push(timerTickMsg)

		case <-s.idleTimer.C:
			if !s.timerMgr.Empty() {
				s.resetIdleTime()
				break
			}
			s.status.Store(idle)

			// check sent after the timeout
			if len(s.mailBox) > 0 {
				s.activate()
			}
			// there are the return just for idle with the mailBox was empty and the timerMgr was empty
			return
		}
	}
}

func (s *actor) handleMsg(msg actor_msg.Message) {
	var message = msg.Message()
	// message is ActorMessage when msg from remote OnRecv
	if actMsg, ok := message.(*actor_msg.ActorMessage); ok {
		if actMsg.Data != nil && actMsg.MsgName != "" {
			defer msg.Free()
			message = actMsg.Fill(s.system.protoIndex)
		}
	}
	if message == nil {
		return
	}

	// recording slow processed of message
	beginTime := tools.Milliseconds()
	defer func() {
		endTime := tools.Milliseconds()
		dur := endTime - beginTime
		if dur > int64(100*time.Millisecond) {
			log.SysLog.Warnw("too long to process time", "msg", msg)
		}
	}()

	reqSourceId, _, _, reqOK := ParseRequestId(msg.GetRequestId())
	if reqOK {
		//recv Response
		if s.id == reqSourceId {
			s.doneRequest(msg.GetRequestId(), message)
			return
		}
		// recv Request
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
	nextAt := s.timerMgr.NextAt()
	if nextAt > 0 {
		return
	}

	d := tools.Maxi64(nextAt-tools.NowTime(), 1)
	if d > 0 && d != s.nextAt {
		if !s.timer.Stop() {
			select {
			case <-s.timer.C:
			default:
			}
		}
		s.timer.Reset(time.Duration(d))

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

	s.system.DispatchEvent(s.id, EvDelactor{ActorId: s.id, Publish: s.remote})
	s.system.actorCache.Delete(s.ID())
	s.system.waitStop.Done()
}

func (s *actor) RegistCmd(cmd string, fn func(...string), usage ...string) {
	if s.system.cmd != nil {
		s.system.cmd.RegistCmd(s.id, cmd, fn, usage...)
	}
}

// Extra Option //////////

func SetMailBoxSize(boxSize int) Option {
	return func(a *actor) {
		a.mailBox = make(chan actor_msg.Message, boxSize)
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
		a.RegistCmd("loadlua", func(s ...string) {
			a.lua.Load(path)
		}, "load lua")
	}
}
