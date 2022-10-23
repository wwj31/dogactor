package actor

import (
	"fmt"
	"github.com/wwj31/dogactor/expect"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

/*
	actor system
*/

const (
	DefaultSysAddr     = ":8888"
	DefaultProfileAddr = ":8760"
)

type SystemOption func(*System) error

type System struct {
	CStop chan struct{}

	sysAddr     string          // cluster listen addr
	profileAddr string          // profile listen addr
	waitStop    *sync.WaitGroup // stop wait
	exiting     int32           // state of stopping

	actorCache sync.Map    // all local actor
	newList    chan *actor // new list

	cluster *actor

	protoIndex *tools.ProtoIndex

	evDispatcher
}

func NewSystem(op ...SystemOption) (*System, error) {
	s := &System{
		sysAddr:     DefaultSysAddr,
		profileAddr: DefaultProfileAddr,
		CStop:       make(chan struct{}, 1),
		waitStop:    &sync.WaitGroup{},
		newList:     make(chan *actor, 100),
	}
	s.evDispatcher = newEvent(s)

	for _, f := range op {
		if e := f(s); e != nil {
			return nil, fmt.Errorf("%w %v", actorerr.ActorSystemOptionErr, e.Error())
		}
	}
	if &s.protoIndex == nil {
		log.SysLog.Warnw("without protobuf index,can't find ptoro struct")
	}

	// SystemOption maybe register actor
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

	// start pprof http server
	profileHttp(s)

	log.SysLog.Infof("System Start")
	return s, nil
}

func (s *System) Stop() {
	if atomic.CompareAndSwapInt32(&s.exiting, 0, 1) {
		go func() {
			// notify all actor to stop
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

// Spawn create a new actor with startup
func (s *System) Spawn(id string, handler spawnActor, op ...Option) (Actor, error) {
	newActor := New(id, handler, op...)
	if err := s.Add(newActor); err != nil {
		return nil, err
	}
	return newActor, nil
}

// Add startup a new actor
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
func (s *System) Send(sourceId, targetId, requestId string, msg interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errFormat(err, sourceId, targetId, requestId, reflect.TypeOf(msg).String())
		}
	}()

	var atr *actor
	if localActor, ok := s.actorCache.Load(targetId); ok {
		atr = localActor.(*actor)
	} else {
		pt, canRemote := msg.(proto.Message)
		if canRemote {
			atr = s.Cluster()
			bytes, marshalErr := proto.Marshal(pt)
			if marshalErr != nil {
				return fmt.Errorf("%w %v", actorerr.ProtoMarshalErr, err)
			}

			// remote message
			actorMsg := &actor_msg.ActorMessage{
				SourceId:  sourceId,
				TargetId:  targetId,
				RequestId: requestId,
				MsgName:   s.protoIndex.MsgName(pt),
				Data:      bytes,
			}

			msg, _ = actorMsg.Marshal()
		} else {
			return actorerr.ActorMsgTypeCanNotRemoteErr
		}
	}

	if atr == nil {
		return actorerr.ActorNotFoundErr
	}

	localMsg := actor_msg.NewActorMessage() // local message
	localMsg.SourceId = sourceId
	localMsg.TargetId = targetId
	localMsg.RequestId = requestId
	localMsg.SetMessage(msg)

	return atr.push(localMsg)
}

// RequestWait sync request
func (s *System) RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (resp interface{}, err error) {
	var t time.Duration
	if len(timeout) > 0 && timeout[0] > 0 {
		t = timeout[0]
	}

	waitRsp := make(chan result)
	waiter := New(
		"wait_"+tools.XUID(),
		&waitActor{c: waitRsp, msg: msg, targetId: targetId, timeout: t},
		//SetLocalized(),
	)
	expect.Nil(s.Add(waiter))

	// wait to result
	r := <-waitRsp
	return r.result, r.err
}

func (s *System) LocalActor(actorId string) *actor {
	v, ok := s.actorCache.Load(actorId)
	if ok {
		return v.(*actor)
	}
	return nil
}

// if ok != nil, caller wait for actor call init to finish
func (s *System) runActor(actor *actor, ok chan<- struct{}) {
	if atomic.LoadInt32(&s.exiting) == 1 && !actor.isWaitActor() {
		return
	}

	go actor.init(ok)
}

func (s *System) Address() string {
	return s.sysAddr
}
func (s *System) ProfileAddr() string {
	return s.profileAddr
}

func (s *System) SetCluster(act *actor) {
	s.cluster = act
}

func (s *System) Cluster() *actor {
	return s.cluster
}

func (s *System) ProtoIndex() *tools.ProtoIndex {
	return s.protoIndex
}

// ProtoIndex index proto struct
func ProtoIndex(pi *tools.ProtoIndex) SystemOption {
	return func(system *System) error {
		system.protoIndex = pi
		return nil
	}
}

func Addr(addr string) SystemOption {
	return func(system *System) error {
		system.sysAddr = addr
		return nil
	}
}

func ProfileAddr(addr string) SystemOption {
	return func(system *System) error {
		system.profileAddr = addr
		return nil
	}
}

func errFormat(err error, sourceId, targetId, requestId, msg string) error {
	return fmt.Errorf("%w s:[%v] t:[%v],r:[%v] msg:[%v]", err, sourceId, targetId, requestId, msg)
}
