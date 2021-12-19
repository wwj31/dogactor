package actor

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/tools"
)

/*
所有actor的驱动器和调度器
*/

type SystemOption func(*System) error

type System struct {
	CStop chan struct{}

	actorAddr string          // 远程actor连接端口
	waitStop  *sync.WaitGroup // stop wait
	exiting   int32           // 停止标记

	// 本地actor管理
	actorCache sync.Map    // 所有本地actor
	newList    chan *actor // 等待启动的actor列表

	//集群管理actorId
	clusterId string

	// 辅助模块
	cmd ICmd
	evDispatcher
}

func NewSystem(op ...SystemOption) (*System, error) {
	s := &System{
		CStop:    make(chan struct{}, 1),
		waitStop: &sync.WaitGroup{},
		newList:  make(chan *actor, 100),
	}
	s.evDispatcher = newEvent(s)

	for _, f := range op {
		if e := f(s); e != nil {
			return nil, fmt.Errorf("%w %w", actorerr.ActorSystemOptionErr, e.Error())
		}
	}

	// SystemOption maybe regist actor
	for len(s.newList) > 0 {
		cluster := <-s.newList
		wait := make(chan struct{})
		s.runActor(cluster, wait)
		<-wait
	}

	go func() {
		for {
			select {
			case actor, ok := <-s.newList:
				if !ok {
					return
				}
				s.runActor(actor, nil)
			}
		}
	}()
	logger.Info("System Start")
	return s, nil
}

func (s *System) runActor(actor *actor, ok chan struct{}) {
	if atomic.LoadInt32(&s.exiting) == 1 && !actor.isWaitActor() {
		return
	}

	go func() {
		actor.run(ok)
		// exit
		logger.KV("actor", actor.ID()).Info("actor done")
		s.actorCache.Delete(actor.ID())
		s.waitStop.Done()
	}()
}

func (s *System) Address() string {
	return s.actorAddr
}
func (s *System) SetCluster(id string) {
	s.clusterId = id
}

func (s *System) Stop() {
	if atomic.CompareAndSwapInt32(&s.exiting, 0, 1) {
		go func() {
			//shutdown() 通知所有actor执行关闭
			s.actorCache.Range(func(key, value interface{}) bool {
				value.(*actor).stop()
				return true
			})
			if s.clusterId != "" {
				for c := false; !c; {
					s.actorCache.Range(func(key, value interface{}) bool {
						c = key == s.clusterId
						return c
					})
					runtime.Gosched()
				}
			}

			e := s.Send("", s.clusterId, "", "stop")
			if e != nil {
				logger.KV("actorerr", e).Error("stop error")
			}
			s.waitStop.Wait()
			close(s.newList)
			logger.Info("System Exit")
			s.CStop <- struct{}{}
		}()
	}
}

// Regist 注册actor，外部创建对象，保证ActorId唯一性
func (s *System) Regist(actor *actor) error {
	if atomic.LoadInt32(&s.exiting) == 1 && !actor.isWaitActor() {
		return fmt.Errorf("%w actor:%v", actorerr.RegisterActorSystemErr, actor.ID())
	}

	actor.system = s

	if _, has := s.actorCache.LoadOrStore(actor.ID(), actor); has {
		return fmt.Errorf("%w actor:%v", actorerr.RegisterActorSameIdErr, actor.ID())
	}

	s.waitStop.Add(1)
	tools.Try(
		func() { s.newList <- actor },
		func(ex interface{}) {
			s.waitStop.Done()
			s.actorCache.Delete(actor.ID())
		})
	return nil
}

// Send actor之间发送消息,
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
		return errFormat(actorerr.ActorNotFoundErr, sourceId, targetId, requestId)
	}

	localMsg := actor_msg.NewLocalActorMessage(sourceId, targetId, requestId, msg)
	if err := atr.push(localMsg); err != nil {
		return errFormat(err, sourceId, targetId, requestId)
	}
	return nil
}

// Addr 设置Actor监听的端口
func Addr(addr string) SystemOption {
	return func(system *System) error {
		system.actorAddr = addr
		return nil
	}
}

func errFormat(err error, sourceId, targetId, requestId string) error {
	return fmt.Errorf("%w s[%v] t[%v],r[%v]", err, sourceId, targetId, requestId)
}
