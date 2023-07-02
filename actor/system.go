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
	"github.com/wwj31/dogactor/actor/internal/innermsg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/logger"
	"github.com/wwj31/dogactor/tools"
)

/*
	actor system
*/

type System struct {
	Stopped chan struct{}

	// bind the inter-service communication port
	// to system-assigned random port
	Addr string

	name          string
	waitStop      *sync.WaitGroup // stop wait
	stopping      atomic.Bool     // state of stopping
	actorCache    sync.Map        // all local actor
	newList       chan *actor     // newcomers
	clusterId     Id              // exactly one of etcd(based service discovery) or mq
	requestWaiter string          // implement sync request
	protoIndex    *tools.ProtoIndex
	evDispatcher
}

func NewSystem(op ...SystemOption) (*System, error) {
	s := &System{
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

	if s.protoIndex == nil {
		log.SysLog.Warnw("without protobuf index,can't find ptoro struct")
	}

	s.requestWaiter = "waiter_" + s.name + "_" + tools.XUID()
	_ = s.NewActor(s.requestWaiter, &waiter{}, SetLocalized())

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

	return s, nil
}

func (s *System) Stop() {
	if s.stopping.CompareAndSwap(false, true) {
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

	newer.appendHandler(func(message Message) bool {
		handler.OnHandle(message)
		return false
	})

	newer.appendHandler(func(message Message) bool {
		if fn, ok := message.Payload().(func()); ok {
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
	if s.stopping.Load() {
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
	var sendToCluster bool
	if atr == nil {
		sendToCluster = true
	} else {
		if drainer, ok := atr.handler.(Drainer); ok {
			sendToCluster = drainer.isDraining()
		}
	}

	if sendToCluster {
		pt, canRemote := msg.(proto.Message)
		if canRemote {
			v, _ := s.actorCache.Load(s.Cluster())
			atr = v.(*actor)
			bytes, marshalErr := proto.Marshal(pt)
			if marshalErr != nil {
				return fmt.Errorf("%w %v", actorerr.ProtoMarshalErr, err)
			}

			// remote message
			actorMsg := &innermsg.ActorMessage{
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

	localMsg := innermsg.NewActorMessage() // local message
	localMsg.SourceId = sourceId
	localMsg.TargetId = targetId
	localMsg.RequestId = requestId.String()
	localMsg.SetPayload(msg)

	return atr.push(localMsg)
}

func (s *System) HasActor(actorId string) bool {
	_, ok := s.actorCache.Load(actorId)
	return ok
}

// if ok != nil, caller wait for actor call init to finish
func (s *System) runActor(actor *actor, ok chan<- struct{}) {
	if s.stopping.Load() {
		return
	}

	go actor.init(ok)
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

func Name(name string) SystemOption {
	return func(system *System) error {
		system.name = name
		return nil
	}
}

func LogFileName(path, fileName string) SystemOption {
	return func(system *System) error {
		log.SysLogOption.LogPath = path
		log.SysLogOption.FileName = fileName
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
