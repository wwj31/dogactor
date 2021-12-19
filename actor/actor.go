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
	"go.uber.org/atomic"
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

		// asyncStop 表示actor外部触发 stop,
		// syncStop 表示actor主动执行 Exit
		asyncStop atomic.Bool
		syncStop  atomic.Bool

		//timerAccuracy 决定计时器更新精度
		timerMgr      jtimer.TimerMgr
		timerAccuracy int64

		//lua
		lua     script.ILua
		luapath string

		//logger
		logger log.Logger

		//request记录发送出去的所有请求，收到返回后，从map删掉
		requests map[string]*request
	}
)

// New 创建actor
// id 		actorId外部定义  @和$为内部符号，其他id尽量不占用
// handler  消息处理模块
// op  修改默认属性
func New(id string, handler spawnActor, op ...Option) *actor {
	a := &actor{
		id:            id,
		handler:       handler,
		mailBox:       make(chan actor_msg.IMessage, 100),
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
func (s *actor) ID() string {
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

// AddTimer 添加计时器,每个actor独立一个计时器
// timeId        计时器id,一般传UUID
// interval 	 单位nanoseconds
// trigger_times 执行次数 -1 无限次
// callback 	 只能是主线程回调
func (s *actor) AddTimer(timeId string, interval time.Duration, callback func(dt int64), trigger_times ...int32) string {
	if int64(interval) < s.timerAccuracy {
		interval = time.Duration(s.timerAccuracy)
	}

	now := tools.Now().UnixNano()
	times := int32(1)
	if len(trigger_times) > 0 {
		times = trigger_times[0]
	}
	newTimer, e := jtimer.NewTimer(now, now+interval.Nanoseconds(), times, callback, timeId)
	if e != nil {
		s.logger.KV("error", e).ErrorStack(3, "AddTimer failed")
		return ""
	}
	return s.timerMgr.AddTimer(newTimer)
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
		s.logger.KVs(log.Fields{"len": l, "cap": c, "actor": s.id}).WarnStack(2, "mail box will quickly full")
	}

	s.mailBox <- msg
	return nil
}

// 定期tick，执行定时器
var timerTickMsg = &actor_msg.ActorMessage{}

func (s *actor) run(ok chan struct{}) {
	s.logger.Debug("actor startup")

	tools.Try(s.handler.OnInit)
	if ok != nil {
		ok <- struct{}{}
	}

	_ = s.system.DispatchEvent(s.id, &EvNewactor{ActorId: s.id, Publish: s.remote})
	defer func() { _ = s.system.DispatchEvent(s.id, &EvDelactor{ActorId: s.id, Publish: s.remote}) }()

	upTimer := time.NewTicker(time.Millisecond * time.Duration(s.timerAccuracy))
	defer upTimer.Stop()

	for {
		if s.stopCheck() {
			return
		}

		select {
		case <-upTimer.C:
			if !s.asyncStop.Load() {
				_ = s.push(timerTickMsg)
			}
		case msg := <-s.mailBox:
			tools.Try(func() {
				s.handleMsg(msg)
				s.timerMgr.Update(tools.Nanoseconds())
			}, nil)
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
	if message.Message() == nil {
		return
	}

	// 慢日志记录
	beginTime := tools.Milliseconds()
	defer func() {
		endTime := tools.Milliseconds()
		dur := endTime - beginTime
		if dur > int64(200*time.Millisecond) {
			s.logger.KV("message", msg).KV("dur", dur).Warn("handle too long")
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

func SetTimerAccuracy(acc int64) Option {
	return func(a *actor) {
		a.timerAccuracy = acc
	}
}

func SetLocalized() Option {
	return func(a *actor) {
		a.remote = false
	}
}

func SetLua(path string) Option {
	expect.True(path != "", log.Fields{"path": "lua path is error"})
	return func(a *actor) {
		a.lua = script.New()
		a.register2Lua()
		a.luapath = path
		a.lua.Load(a.luapath)
		a.RegistCmd("loadlua", func(s ...string) {
			a.lua.Load(path)
		}, "加载lua脚本")
	}
}

//=========================
