package actor

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/logger"
	"github.com/wwj31/dogactor/tools"
)

/*
	actor system
*/

const (
	DefaultSysAddr = ":8888"
)

type System struct {
	Stopped chan struct{}

	name          string
	sysAddr       string          // cluster listen addr
	waitStop      *sync.WaitGroup // stop wait
	exiting       int32           // state of stopping
	actorCache    sync.Map        // all local actor
	newList       chan *actor     // newcomers
	clusterId     Id              // exactly one of etcd(based service discovery) or mq
	requestWaiter string          // implement sync request
	protoIndex    *tools.ProtoIndex
	evDispatcher
}

func NewSystem(op ...SystemOption) (*System, error) {
	s := &System{
		sysAddr:  DefaultSysAddr,
		Stopped:  make(chan struct{}, 1),
		waitStop: &sync.WaitGroup{},
		newList:  make(chan *actor, 100),
	}
	s.evDispatcher = newEvent(s)

	for _, f := range op {
		if e := f(s); e != nil {
			return nil, fmt.Errorf("%w %v", actorerr.ActorSystemOptionErr, e.Error())
		}
	}
	log.Init()

	if &s.protoIndex == nil {
		log.SysLog.Warnw("without protobuf index,can't find ptoro struct")
	}

	s.requestWaiter = "waiter_" + s.name + "_" + tools.XUID()
	_ = s.NewActor(s.requestWaiter, &waiter{})

	// first,create waiter and cluster
	for len(s.newList) > 0 {
		wait := make(chan struct{})
		s.runActor(<-s.newList, wait)
		<-wait
	}

	go func() {
		for {
			select {
			case newcomer, ok := <-s.newList:
				if !ok {
					return
				}
				s.runActor(newcomer, nil)
			}
		}
	}()

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

			var _stop bool
			for !_stop {
				var actorId string

				s.actorCache.Range(func(key, value interface{}) bool {
					if key == s.clusterId || key == s.requestWaiter {
						return true
					}

					actorId = key.(string)
					log.SysLog.Warnw("spare actor...", "actorId", actorId)
					return false
				})

				if actorId == "" {
					_stop = true
				}

				time.Sleep(time.Second)
				runtime.Gosched()
			}

			if err := s.Send("", s.requestWaiter, "", "stop"); err != nil {
				log.SysLog.Errorw("stop request waiter send failed", "err", err)
			}

			s.waitStop.Wait()
			close(s.newList)
			log.SysLog.Infof("System Exit")
			s.Stopped <- struct{}{}
		}()
	}
}

// NewActor new an actor for system
// id is invalid if contain '@' or '$'
func (s *System) NewActor(id Id, handler spawnActor, opt ...Option) error {
	newer := &actor{
		system:   s,
		id:       id,
		handler:  handler,
		msgChain: make([]func(message Message) bool, 0, 1),
		mailBox:  mailBox{ch: make(chan Message, 100)},
		remote:   true,
	}

	newer.AppendHandler(func(message Message) bool {
		handler.OnHandle(message)
		return false
	})

	newer.AppendHandler(func(message Message) bool {
		if fn, ok := message.RawMsg().(func()); ok {
			fn()
			return false
		}
		return true
	})

	newer.status.Store(starting)

	handler.initActor(newer)

	for _, f := range opt {
		f(newer)
	}
	return s.add(newer)
}

// Add startup a new actor
func (s *System) add(actor *actor) error {
	if atomic.LoadInt32(&s.exiting) == 1 {
		return fmt.Errorf("%w actor:%v", actorerr.RegisterActorSystemErr, actor.ID())
	}

	if _, has := s.actorCache.LoadOrStore(actor.ID(), actor); has {
		return fmt.Errorf("%w actor:%v", actorerr.RegisterActorSameIdErr, actor.ID())
	}

	s.waitStop.Add(1)
	var err error
	tools.Try(
		func() { s.newList <- actor },
		func(ex interface{}) {
			s.waitStop.Done()
			s.actorCache.Delete(actor.ID())
			err = fmt.Errorf("add actor failed with exception:%v ", ex)
		})
	return err
}

// Send try to send the message to target
func (s *System) Send(sourceId, targetId Id, requestId RequestId, msg any) (err error) {
	defer func() {
		if err != nil {
			err = errFormat(err, sourceId, targetId, requestId, reflect.TypeOf(msg).String())
		}
	}()

	var atr *actor
	if localActor, ok := s.actorCache.Load(targetId); ok {
		atr = localActor.(*actor)
	}

	// when the actor into draining mode or not found locally,
	// send the message to the cluster.
	if drainer, ok := atr.handler.(Drainer); (ok && drainer.Draining()) || atr == nil {
		pt, canRemote := msg.(proto.Message)
		if canRemote {
			v, _ := s.actorCache.Load(s.Cluster())
			atr = v.(*actor)
			bytes, marshalErr := proto.Marshal(pt)
			if marshalErr != nil {
				return fmt.Errorf("%w %v", actorerr.ProtoMarshalErr, err)
			}

			// remote message
			actorMsg := &actor_msg.ActorMessage{
				SourceId:  sourceId,
				TargetId:  targetId,
				RequestId: requestId.String(),
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
	localMsg.RequestId = requestId.String()
	localMsg.SetMessage(msg)

	return atr.push(localMsg)
}

func (s *System) HasActor(actorId string) bool {
	_, ok := s.actorCache.Load(actorId)
	return ok
}

// if ok != nil, caller wait for actor call init to finish
func (s *System) runActor(actor *actor, ok chan<- struct{}) {
	if atomic.LoadInt32(&s.exiting) == 1 {
		return
	}

	go actor.init(ok)
}

func (s *System) Address() string {
	return s.sysAddr
}

func (s *System) WaiterId() string {
	return s.requestWaiter
}

func (s *System) SetCluster(id Id) {
	s.clusterId = id
}

func (s *System) Cluster() Id {
	return s.clusterId
}

func (s *System) ProtoIndex() *tools.ProtoIndex {
	return s.protoIndex
}

type SystemOption func(*System) error

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

func Name(name string) SystemOption {
	return func(system *System) error {
		system.name = name
		return nil
	}
}

func LogLevel(level logger.Level) SystemOption {
	return func(system *System) error {
		log.SysLogOption.Level = level
		return nil
	}
}

func Output(out io.Writer) SystemOption {
	return func(system *System) error {
		log.SysLogOption.ExtraWriter = append(log.SysLogOption.ExtraWriter, out)
		log.SysLogOption.LogPath = ""
		return nil
	}
}

func errFormat(err error, sourceId, targetId Id, requestId RequestId, msg string) error {
	return fmt.Errorf("%w s:[%v] t:[%v],r:[%v] msg:[%v]", err, sourceId, targetId, requestId, msg)
}
