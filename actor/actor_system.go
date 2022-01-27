package actor

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/wwj31/dogactor/log"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/tools"
)

/*
	actor system
*/

type SystemOption func(*System) error

type System struct {
	CStop chan struct{}

	actorAddr string          // cluster listen addr
	waitStop  *sync.WaitGroup // stop wait
	exiting   int32           // state of stopping

	actorCache sync.Map    // all of local actor
	newList    chan *actor // new list

	cluster *actor

	// extra
	cmd Cmder
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
			return nil, fmt.Errorf("%w %v", actorerr.ActorSystemOptionErr, e.Error())
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
	log.SysLog.Infof("System Start")
	return s, nil
}

// if ok != nil, caller wait to actor inited
func (s *System) runActor(actor *actor, ok chan<- struct{}) {
	if atomic.LoadInt32(&s.exiting) == 1 && !actor.isWaitActor() {
		return
	}

	go func() {
		actor.run(ok)
		// exit
		log.SysLog.Infow("actor done", "actorId", actor.ID())
		s.actorCache.Delete(actor.ID())
		s.waitStop.Done()
	}()
}

func (s *System) Address() string {
	return s.actorAddr
}
func (s *System) SetCluster(act *actor) {
	s.cluster = act
}

func (s *System) Stop() {
	if atomic.CompareAndSwapInt32(&s.exiting, 0, 1) {
		go func() {
			// notify all of actor to stop
			s.actorCache.Range(func(key, value interface{}) bool {
				value.(*actor).stop()
				return true
			})
			if s.cluster != nil {
				for c := false; !c; {
					s.actorCache.Range(func(key, value interface{}) bool {
						c = (key.(string)) == s.cluster.id
						return c
					})
					runtime.Gosched()
				}
			}

			if s.cluster != nil {
				e := s.Send("", s.cluster.id, "", "stop")
				if e != nil {
					log.SysLog.Errorw("system stop exception", "err", e)
				}
			}

			s.waitStop.Wait()
			close(s.newList)
			log.SysLog.Infof("System Exit %v", s.Address())
			s.CStop <- struct{}{}
		}()
	}
}

func (s *System) Add(actor *actor) error {
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

// Send msg send to target,if target not exist in local cache msg shall send to cluster
func (s *System) Send(sourceId, targetId, requestId string, msg interface{}) error {
	var atr *actor
	if localActor, ok := s.actorCache.Load(targetId); ok {
		atr = localActor.(*actor)
	} else {
		pt, canRemote := msg.(proto.Message)
		if canRemote {
			atr = s.cluster
			bytes, err := proto.Marshal(pt)
			if err != nil {
				return errFormat(fmt.Errorf("%w %v", actorerr.ProtoMarshalErr, err), sourceId, targetId, requestId)
			}
			actorMsg := &actor_msg.ActorMessage{} // remote message
			actorMsg.SourceId = sourceId
			actorMsg.TargetId = targetId
			actorMsg.RequestId = requestId
			actorMsg.MsgName = tools.MsgName(pt)
			actorMsg.Data = bytes
			msg, _ = actorMsg.Marshal()
		}
	}

	if atr == nil {
		return errFormat(actorerr.ActorNotFoundErr, sourceId, targetId, requestId)
	}

	localMsg := actor_msg.NewActorMessage() // local message
	localMsg.SourceId = sourceId
	localMsg.TargetId = targetId
	localMsg.RequestId = requestId
	localMsg.SetMessage(msg)

	if err := atr.push(localMsg); err != nil {
		return errFormat(err, sourceId, targetId, requestId)
	}
	return nil
}

func Addr(addr string) SystemOption {
	return func(system *System) error {
		system.actorAddr = addr
		return nil
	}
}

func errFormat(err error, sourceId, targetId, requestId string) error {
	return fmt.Errorf("%w s[%v] t[%v],r[%v]", err, sourceId, targetId, requestId)
}
