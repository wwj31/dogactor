package actor

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
	"github.com/wwj31/godactor/actor/err"
	"github.com/wwj31/godactor/actor/internal/actor_msg"
	"github.com/wwj31/godactor/tools"
)

/* 所有actor的驱动器和调度器*/
type SystemOption func(*System) error

type System struct {
	actorAddr string          // 远程actor连接端口
	waitStop  *sync.WaitGroup // stop wait
	exiting   int32           // 停止标记

	// 本地actor管理
	actorCache sync.Map    // 所有本地actor
	waitRun    chan *actor // 等待启动的actor列表

	//集群管理actorId
	clusterId string

	// 辅助模块
	cmd ICmd
	*evDispatcher
}

func NewSystem(op ...SystemOption) (*System, error) {
	sys := &System{
		waitStop: &sync.WaitGroup{},
		waitRun:  make(chan *actor, 100),
	}
	sys.evDispatcher = newEvent(sys)

	for _, f := range op {
		if e := f(sys); e != nil {
			return nil, fmt.Errorf("%w %w", err.ActorSystemOptionErr, e.Error())
		}
	}

	for len(sys.waitRun) > 0 {
		cluster := <-sys.waitRun
		wait := make(chan struct{})
		sys.runActor(cluster, wait)
		<-wait
	}

	tools.GoEngine(func() {
		for {
			select {
			case actor, ok := <-sys.waitRun:
				if !ok {
					return
				}
				sys.runActor(actor, nil)
			}
		}
	})
	logger.Info("System Start")
	return sys, nil
}
func (s *System) waitCluster() {
	for {
		continueWait := false
		s.actorCache.Range(func(key, value interface{}) bool {
			if continueWait = key != s.clusterId; continueWait {
				return false
			}
			return true
		})

		if !continueWait {
			return
		}
		runtime.Gosched()
	}
}

func (s *System) Address() string {
	return s.actorAddr
}
func (s *System) SetCluster(id string) {
	s.clusterId = id
}

func (s *System) Stop() {
	if atomic.CompareAndSwapInt32(&s.exiting, 0, 1) {
		//shutdown() 通知所有actor执行关闭
		s.actorCache.Range(func(key, value interface{}) bool {
			value.(*actor).stop()
			return true
		})
		s.waitCluster()
		s.Send("", s.clusterId, "", "stop")
		s.waitStop.Wait()
		close(s.waitRun)
		logger.Info("System Exit")
	}
}

// 注册actor，外部创建对象，保证ActorId唯一性
func (s *System) Regist(actor *actor) error {
	if atomic.LoadInt32(&s.exiting) == 1 && !actor.isWaitActor() {
		return fmt.Errorf("%w actor:%v", err.RegisterActorSystemErr, actor.GetID())
	}

	actor.setSystem(s)
	if _, has := s.actorCache.LoadOrStore(actor.GetID(), actor); has {
		return fmt.Errorf("%w actor:%v", err.RegisterActorSameIdErr, actor.GetID())
	}

	s.waitStop.Add(1)
	tools.Try(
		func() { s.waitRun <- actor },
		func(ex interface{}) {
			s.waitStop.Done()
			s.actorCache.Delete(actor.GetID())
		})
	return nil
}

func (s *System) runActor(actor *actor, ok chan struct{}) {
	if atomic.LoadInt32(&s.exiting) == 1 && !actor.isWaitActor() {
		return
	}

	go func() {
		actor.run(ok)
		// exit
		logger.KV("actor", actor.GetID()).Info("actor done")
		s.actorCache.Delete(actor.GetID())
		s.waitStop.Done()
	}()
}

// actor之间发送消息,
// sourceid 发送源actor
// targetid 目标actor
// message 消息内容
func (s *System) Send(sourceId, targetId, requestId string, msg interface{}) error {
	var atr *actor
	if localActor, ok := s.actorCache.Load(targetId); ok { //消息能否发给本地
		atr = localActor.(*actor)
	} else {
		_, canRemote := msg.(proto.Message)
		cluster, ok := s.actorCache.Load(s.clusterId)
		if canRemote && ok { // 消息能否发送给远端
			atr = cluster.(*actor)
		}
	}

	if atr == nil {
		return fmt.Errorf("%w s[%v] t[%v],r[%v]", err.ActorNotFoundErr, sourceId, targetId, requestId)
	}

	if e := atr.push(actor_msg.NewLocalActorMessage(sourceId, targetId, requestId, msg)); e != nil {
		return fmt.Errorf("%w s[%v] t[%v],r[%v]", e, sourceId, targetId, requestId)
	}
	return nil
}

func (s *System) ClusterId() string {
	return s.clusterId
}

// 设置Actor监听的端口
func Addr(addr string) SystemOption {
	return func(system *System) error {
		system.actorAddr = addr
		return nil
	}
}
