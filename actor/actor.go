package actor

import (
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/internal/script"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"github.com/wwj31/jtimer"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/atomic"
)

type (
	Actor interface {
		//core
		GetID() string
		System() *System
		Exit()

		//timer
		AddTimer(timeId string, interval time.Duration, callback func(dt int64), count ...int32) string
		CancelTimer(timerId string)

		//lua
		CallLua(name string, ret int, args ...lua.LValue) []lua.LValue

		// Send 不保证消息发送可靠性，如果需要超时处理用 Request
		Send(targetId string, msg interface{})
		Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request)
		RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (result interface{}, err error)
		Response(requestId string, msg interface{}) error

		//cmd
		RegistCmd(cmd string, fn func(...string))
	}

	// actor 处理接口
	actorHandler interface {
		onInitActor(actor Actor)

		OnInit()
		OnStop() bool // true 立刻停止，false 延迟停止

		OnHandleEvent(event interface{})                                             //事件消息
		OnHandleMessage(sourceId, targetId string, msg interface{})                  //普通消息
		OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) error //请求消息(需要应答)
	}
)
type (
	ActorOption func(*actor)
	// 运行单元
	actor struct {
		id        string
		handler   actorHandler
		mailBox   chan actor_msg.IMessage // 这里的chan是 多生产者单消费者模式，不需要close，正常退出即可
		system    *System
		remote    bool // 是否能被远端发现 默认为true, 如果是本地actor,手动设SetLocalized()后,再注册
		asyncStop atomic.Bool
		syncStop  atomic.Bool

		//timer
		timerMgr      *jtimer.TimerMgr
		timerAccuracy int64 // 计时器精度

		//lua
		lua     script.ILua
		luapath string

		//logger
		logger *log.Logger

		//request
		requests map[string]*request
	}
)

// 创建actor
// id 		actorId外部定义  @和$为内部符号，其他id尽量不占用
// handler  消息处理模块
// op  修改默认属性
func New(id string, handler actorHandler, op ...ActorOption) *actor {
	a := &actor{
		id:            id,
		handler:       handler,
		mailBox:       make(chan actor_msg.IMessage, 1000),
		remote:        true, // 默认都能被远端发现
		timerMgr:      jtimer.NewTimerMgr(),
		timerAccuracy: 500,
		logger:        log.NewWithDefaultAndLogger(logger, map[string]interface{}{"actorId": id}),
		requests:      make(map[string]*request),
	}

	handler.onInitActor(a)

	for _, f := range op {
		f(a)
	}
	return a
}

// 获取actorID
func (s *actor) GetID() string {
	return s.id
}

func (s *actor) isWaitActor() bool {
	_, ok := s.handler.(*waitActor)
	return ok
}

// 业务层触发关闭
func (s *actor) Exit() {
	s.syncStop.Swap(true)
}

// system触发关闭
func (s *actor) stop() {
	s.asyncStop.Swap(true)
}

// 添加计时器,每个actor独立一个计时器
// timeId        计时器id,一般传UUID
// interval 	 单位nanoseconds
// count         执行次数 -1 无限次
// callback 	 只能是主线程回调
func (s *actor) AddTimer(timeId string, interval time.Duration, callback func(dt int64), count ...int32) string {
	var (
		now   = tools.Now().UnixNano()
		endAt = now + interval.Nanoseconds()
		times = int32(0)
	)

	if int64(interval) < s.timerAccuracy {
		interval = time.Duration(s.timerAccuracy)
	}

	upErr := s.timerMgr.UpdateTimer(timeId, endAt)

	if jtimer.ErrorUpdateTimer == upErr {
		times = int32(1)
		if len(count) > 0 {
			times = count[0]
		}

		newTimer, e := jtimer.NewTimer(now, endAt, times, callback, timeId)
		if e != nil {
			s.logger.KV("error", e).ErrorStack(3, "AddTimer failed")
			return ""
		}
		s.timerMgr.AddTimer(newTimer)
	}
	return timeId
}

// 删除一个定时器
func (s *actor) CancelTimer(timerId string) {
	s.timerMgr.CancelTimer(timerId)
}

// Push一个消息
func (s *actor) push(msg actor_msg.IMessage) error {
	if msg == nil {
		return actorerr.ActorPushMsgErr
	}

	if l, c := len(s.mailBox), cap(s.mailBox); l > c*2/3 {
		s.logger.KVs(log.Fields{"len": l, "cap": c, "actor": s.id}).Warn("mail box will quickly full")
	}

	s.mailBox <- msg
	return nil
}

func (s *actor) run(ok chan struct{}) {
	s.logger.Debug("actor startup")

	tools.Try(func() { s.handler.OnInit() }, nil)
	if ok != nil {
		ok <- struct{}{}
	}

	s.system.DispatchEvent(s.id, &Ev_newActor{ActorId: s.id, Publish: s.remote})
	defer func() { s.system.DispatchEvent(s.id, &Ev_delActor{ActorId: s.id, Publish: s.remote}) }()

	upTimer := time.NewTicker(time.Millisecond * time.Duration(s.timerAccuracy))
	defer upTimer.Stop()

	for {
		if s.stopCheck() {
			return
		}

		select {
		case <-upTimer.C:
			if !s.asyncStop.Load() {
				tools.Try(func() { s.timerMgr.Update(tools.Now().UnixNano()) }, nil)
			}
		case msg := <-s.mailBox:
			tools.Try(func() { s.handleMsg(msg) }, nil)
			msg.Free()
		}
	}
}

func (s *actor) handleMsg(msg actor_msg.IMessage) {
	s.logger.KVs(log.Fields{"message": msg}).Bule().Debug("Recv ActorMessage")
	message, ok := msg.(*actor_msg.ActorMessage)
	if !ok {
		s.logger.KVs(log.Fields{"message": message}).Red().Warn("unkown actor message type")
		return
	}

	beginTime := tools.Milliseconds()
	defer func() {
		endTime := tools.Milliseconds()
		dur := endTime - beginTime
		if dur > int64(5*time.Millisecond) {
			s.logger.KV("message", msg).KV("dur", dur).Warn("handle too long")
		}
	}()

	reqSourceId, _, _, reqOK := ParseRequestId(message.RequestId)
	if reqOK {
		if s.id == reqSourceId { //收到Respone
			s.doneRequest(message.RequestId, message.Message())
		} else {
			// 收到Request
			if err := s.handler.OnHandleRequest(message.SourceId, message.TargetId, message.RequestId, message.Message()); err != nil {
				expect.Nil(s.Response(message.RequestId, &actor_msg.RequestDeadLetter{Err: err.Error()}))
			}
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

func (s *actor) stopCheck() (immediatelyStop bool) {
	//逻辑层说停止=>立即停止
	if s.syncStop.Load() {
		return true
	}

	//系统层说停止=>通知逻辑,并且返回是否立即停止
	if s.asyncStop.Load() {
		tools.Try(func() { immediatelyStop = s.handler.OnStop() }, func(ex interface{}) { immediatelyStop = true })
	}
	return
}

//=========================简化actorSystem调用
func (s *actor) setSystem(actorSystem *System)            { s.system = actorSystem }
func (s *actor) System() *System                          { return s.system }
func (s *actor) Send(targetId string, msg interface{})    { s.system.Send(s.id, targetId, "", msg) }
func (s *actor) RegistCmd(cmd string, fn func(...string)) { s.system.RegistCmd(s.id, cmd, fn) }

//=========================
func SetMailBoxSize(boxSize int) ActorOption {
	return func(a *actor) {
		a.mailBox = make(chan actor_msg.IMessage, boxSize)
	}
}

func SetTimerAccuracy(acc int64) ActorOption {
	return func(a *actor) {
		a.timerAccuracy = acc
	}
}

func SetLocalized() ActorOption {
	return func(a *actor) {
		a.remote = false
	}
}

func SetLua(path string) ActorOption {
	expect.True(path != "", log.Fields{"path": "lua path is error"})
	return func(a *actor) {
		a.lua = script.New()
		a.register2Lua()
		a.luapath = path
		a.lua.Load(a.luapath)
	}
}

//=========================
