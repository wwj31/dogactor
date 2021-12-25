package actor

import (
	"github.com/wwj31/jtimer"
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/internal/script"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type (
	Option func(*actor)
	// 运行单元
	actor struct {
		id      string
		handler actorHandler
		mailBox chan actor_msg.IMessage
		system  *System

		// 是否能被远端发现 默认为true, 如果是本地actor,
		// 手动设SetLocalized()后,再注册
		remote bool

		// timer
		timerMgr jtimer.TimerMgr
		timer    *time.Timer
		nextAt   int64

		//lua
		lua     script.ILua
		luapath string

		//request记录发送出去的所有请求，收到返回后，从map删掉
		requests map[string]*request
	}
)

// New new actor
// id 		actorId外部定义  @和$为内部符号，其他id尽量不占用
// handler  消息处理模块
func New(id string, handler spawnActor, op ...Option) *actor {
	a := &actor{
		id:       id,
		handler:  handler,
		mailBox:  make(chan actor_msg.IMessage, 100),
		remote:   true, // 默认都能被远端发现
		timerMgr: jtimer.NewTimerMgr(),
		requests: make(map[string]*request),
	}

	handler.initActor(a)

	for _, f := range op {
		f(a)
	}
	return a
}

// ID get actorID
func (s *actor) ID() string {
	return s.id
}

func (s *actor) isWaitActor() bool {
	_, ok := s.handler.(*waitActor)
	return ok
}

// 业务层触发关闭
func (s *actor) Exit() {
	_ = s.push(actorStopMsg)
}

// system触发关闭
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

// AddTimer 添加计时器,每个actor独立一个计时器
// timeId        计时器id,一般传UUID
// interval 	 单位nanoseconds
// trigger_times 执行次数 -1 无限次
// callback 	 只能是主线程回调
func (s *actor) AddTimer(timeId string, endAt int64, callback func(dt int64), trigger_times ...int32) string {
	now := tools.NowTime()
	times := int32(1)
	if len(trigger_times) > 0 {
		times = trigger_times[0]
	}

	newTimer, e := jtimer.NewTimer(now, endAt, times, callback, timeId)
	if e != nil {
		log.SysLog.Errorw("AddTimer failed", "err", e)
		return ""
	}

	timerId := s.timerMgr.AddTimer(newTimer)
	s.resetTime()
	return timerId
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
func (s *actor) CancelTimer(timerId string) {
	s.timerMgr.CancelTimer(timerId)
	s.resetTime()
}

// Push一个消息
func (s *actor) push(msg actor_msg.IMessage) error {
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
	return nil
}

var (
	timerTickMsg = &actor_msg.ActorMessage{} // timer tick msg
	actorStopMsg = &actor_msg.ActorMessage{} // actor stop msg
)

func (s *actor) run(ok chan<- struct{}) {
	s.timer = time.NewTimer(0)
	if !s.timer.Stop() {
		<-s.timer.C
	}

	tools.Try(s.handler.OnInit)
	if ok != nil {
		ok <- struct{}{}
	}

	s.system.DispatchEvent(s.id, &EvNewactor{ActorId: s.id, Publish: s.remote})
	defer func() {
		s.system.DispatchEvent(s.id, &EvDelactor{ActorId: s.id, Publish: s.remote})
	}()

	for {
		select {
		case <-s.timer.C:
			_ = s.push(timerTickMsg)
		case msg := <-s.mailBox:
			if s.isStop(msg) {
				return
			}
			tools.Try(func() {
				s.handleMsg(msg)

				s.timerMgr.Update(tools.NowTime())
				s.resetTime()
			})
			msg.Free()
		}
	}
}

func (s *actor) handleMsg(msg actor_msg.IMessage) {
	message, ok := msg.(*actor_msg.ActorMessage)
	if !ok {
		log.SysLog.Warnw("unkown type of the message", "msg", message)
		return
	}
	if message.Message() == nil {
		return
	}

	// 慢日志记录
	beginTime := tools.Milliseconds()
	defer func() {
		endTime := tools.Milliseconds()
		dur := endTime - beginTime
		if dur > int64(100*time.Millisecond) {
			log.SysLog.Warnw("too long to process time", "msg", msg)
		}
	}()

	// 处理Req消息
	reqSourceId, _, _, reqOK := ParseRequestId(message.RequestId)
	if reqOK {
		if s.id == reqSourceId { //收到Respone
			s.doneRequest(message.RequestId, message.Message())
			return
		}
		// 收到Request
		if e := s.handler.OnHandleRequest(message.SourceId, message.TargetId, message.RequestId, message.Message()); e != nil {
			expect.Nil(s.Response(message.RequestId, &actor_msg.RequestDeadLetter{Err: e.Error()}))
		}
		return
	}

	//事件消息
	if event, ok := message.Message().(*actor_msg.EventMessage); ok {
		s.handler.OnHandleEvent(event.ActEvent())
		return
	}

	// 函数消息
	if fn, ok := message.Message().(func()); ok {
		fn()
		return
	}

	s.handler.OnHandleMessage(message.SourceId, message.TargetId, message.Message())
}

func (s *actor) resetTime() {
	nextAt := s.timerMgr.NextAt()
	if nextAt > 0 {
		d := tools.Maxi64(nextAt-tools.NowTime(), 1)
		if d > 0 && d != s.nextAt {
			s.timer.Reset(time.Duration(d))
			s.nextAt = nextAt
		}
	}
}
func (s *actor) isStop(msg actor_msg.IMessage) bool {
	message, ok := msg.(*actor_msg.ActorMessage)
	if ok && message == actorStopMsg {
		return true
	}
	return false
}

//=========================简化actorSystem调用
func (s *actor) System() *System { return s.system }

func (s *actor) Send(targetId string, msg interface{}) error {
	return s.system.Send(s.id, targetId, "", msg)
}
func (s *actor) RegistCmd(cmd string, fn func(...string), usage ...string) {
	if s.system.cmd != nil {
		s.system.cmd.RegistCmd(s.id, cmd, fn, usage...)
	}
}

//=========================
func SetMailBoxSize(boxSize int) Option {
	return func(a *actor) {
		a.mailBox = make(chan actor_msg.IMessage, boxSize)
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

//=========================
