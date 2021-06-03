package actor

import (
	"github.com/wwj31/godactor/actor/err"
	"github.com/wwj31/godactor/actor/internal/actor_msg"
	"github.com/wwj31/godactor/actor/internal/script"
	"github.com/wwj31/godactor/expect"
	"github.com/wwj31/godactor/log"
	"github.com/wwj31/godactor/tools"
	"github.com/wwj31/jtimer"
	lua "github.com/yuin/gopher-lua"
	"sync/atomic"
	"time"
)

type (
	IActor interface {
		//core
		GetID() string
		ActorSystem() *ActorSystem
		LogicStop()

		//timer
		AddTimer(interval time.Duration, trigger_times int32, callback jtimer.FuncCallback) int64
		CancelTimer(timerId int64)

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
	IActorHandler interface {
		initActor(actor IActor)

		Init() error
		Stop() (immediatelyStop bool) // true 立刻停止，false 延迟停止

		HandleEvent(event interface{})                                                       //事件消息
		HandleMessage(sourceId, targetId string, msg interface{})                            //普通消息(只管发送)
		HandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) //请求消息(需要应答)
	}
)
type (
	ActorOption func(*actor)
	// 运行单元
	actor struct {
		id          string
		handler     IActorHandler
		mailBox     chan actor_msg.IMessage // todo 考虑用无锁队列优化
		actorSystem *ActorSystem
		remote      bool  // 是否能被远端发现 默认为true, 如果是本地actor,手动设SetLocalized()后,再注册
		systemStop  int32 // 用于外部通知结束
		logicStop   int32 // 用于内部自己结束	(一些actor的关闭是由其他actor调用,防止core退出时重复close(exist)的情况,故特意添加此变量,用于主动调用SuspendStop())

		//timer
		timerMgr      *jtimer.TimerMgr
		timerAccuracy int64 // 计时器精度

		//lua
		lua     script.ILua
		luapath string

		//log
		logger *log.Logger

		//request
		requests map[string]*request
	}
)

// 创建actor
// id 		actorId外部定义  @和$为内部符号，其他id尽量不占用
// handler  消息处理模块
// op  修改默认属性
func NewActor(id string, handler IActorHandler, op ...ActorOption) *actor {
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

	handler.initActor(a)

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

// 逻辑层主动关闭actor调用此函数
func (s *actor) LogicStop() {
	atomic.CompareAndSwapInt32(&s.logicStop, 0, 1)
}

// 系统层主动关闭actor调用此函数
func (s *actor) SystemStop() {
	atomic.CompareAndSwapInt32(&s.systemStop, 0, 1)
}

// 添加计时器,每个actor独立一个计时器
// interval 	 单位nanoseconds
// trigger_times 执行次数 -1 无限次
// callback 	 只能是主线程回调
func (s *actor) AddTimer(interval time.Duration, trigger_times int32, callback jtimer.FuncCallback) int64 {
	if int64(interval) < s.timerAccuracy {
		interval = time.Duration(s.timerAccuracy)
	}

	now := tools.Now().UnixNano()
	newTimer, e := jtimer.NewTimer(now, now+interval.Nanoseconds(), trigger_times, callback)
	if e != nil {
		s.logger.KV("error", e).ErrorStack(3, "AddTimer failed")
		return -1
	}
	return s.timerMgr.AddTimer(newTimer, false)
}

// 删除一个定时器
func (s *actor) CancelTimer(timerId int64) {
	s.timerMgr.CancelTimer(timerId)
}

// Push一个消息
func (s *actor) push(msg actor_msg.IMessage) error {
	if msg == nil {
		return err.ActorPushMsgErr
	}

	if l, c := len(s.mailBox), cap(s.mailBox); l > c*2/3 {
		s.logger.KVs(log.Fields{"len": l, "cap": c, "actor": s.id}).Warn("mail box will quickly full")
	}

	s.mailBox <- msg
	return nil
}

func (s *actor) runActor() {
	if s.GetID() != s.actorSystem.clusterId {
		s.actorSystem.clusterInit.Wait()
	}

	s.logger.Debug("actor startup")

	var err error
	tools.Try(func() { err = s.handler.Init() }, nil)

	if err != nil {
		s.logger.KV("error", err).Error("actor InitGame failed")
		return
	}

	s.actorSystem.DispatchEvent(s.id, &Ev_newActor{ActorId: s.id, Publish: s.remote})
	defer func() { s.actorSystem.DispatchEvent(s.id, &Ev_delActor{ActorId: s.id, Publish: s.remote}) }()

	if s.GetID() == s.actorSystem.clusterId {
		s.actorSystem.clusterInit.Done()
	}

	up_timer := time.NewTicker(time.Millisecond * time.Duration(s.timerAccuracy))
	defer up_timer.Stop()

	for {
		if s.stopCheck() {
			return
		}

		select {
		case <-up_timer.C:
			if atomic.LoadInt32(&s.systemStop) == 0 {
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
			if err := s.handler.HandleRequest(message.SourceId, message.TargetId, message.RequestId, message.Message()); err != nil {
				expect.Nil(s.Response(message.RequestId, &actor_msg.RequestDeadLetter{Err: err.Error()}))
			}
		}
		return
	}

	//事件消息
	if event, ok := message.Message().(*actor_msg.EventMessage); ok {
		s.handler.HandleEvent(event.ActEvent())
		return
	}

	// 函数消息
	if fn, ok := message.Message().(func()); ok {
		fn()
		return
	}

	s.handler.HandleMessage(message.SourceId, message.TargetId, message.Message())
}

func (s *actor) stopCheck() (immediatelyStop bool) {
	//逻辑层说停止=>立即停止
	if atomic.LoadInt32(&s.logicStop) == 1 {
		return true
	}

	//系统层说停止=>通知逻辑,并且返回是否立即停止
	if atomic.CompareAndSwapInt32(&s.systemStop, 1, 2) {
		tools.Try(func() { immediatelyStop = s.handler.Stop() }, func(ex interface{}) { immediatelyStop = true })
	}
	return
}

//=========================简化actorSystem调用
func (s *actor) setSystem(actorSystem *ActorSystem)       { s.actorSystem = actorSystem }
func (s *actor) ActorSystem() *ActorSystem                { return s.actorSystem }
func (s *actor) Send(targetId string, msg interface{})    { s.actorSystem.Send(s.id, targetId, "", msg) }
func (s *actor) RegistCmd(cmd string, fn func(...string)) { s.actorSystem.RegistCmd(s.id, cmd, fn) }

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
