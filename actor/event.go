package actor

import "github.com/wwj31/godactor/actor/err"

type IEvent interface {
	Init(actorSystem *ActorSystem)
	RegistEvent(actorId string, events ...interface{}) error
	CancelEvent(actorId string, events ...interface{}) error
	CancelAll(actorId string)
	DispatchEvent(sourceId string, event interface{}) error
}

// 设置Actor监听的端口
func WithEvent(event IEvent) SystemOption {
	return func(system *ActorSystem) error {
		system.event = event
		event.Init(system)
		return nil
	}
}

func (s *ActorSystem) RegistEvent(actorId string, events ...interface{}) error {
	if s.event == nil {
		return err.EventHasNotErr
	}
	return s.event.RegistEvent(actorId, events...)
}

func (s *ActorSystem) CancelEvent(actorId string, events ...interface{}) error {
	if s.event != nil {
		return err.EventHasNotErr
	}
	return s.event.CancelEvent(actorId, events...)
}

func (s *ActorSystem) CancelAllEvent(actorId string) error {
	if s.event != nil {
		return err.EventHasNotErr
	}
	s.event.CancelAll(actorId)
	return nil
}

func (s *ActorSystem) DispatchEvent(sourceId string, event interface{}) error {
	if s.event != nil {
		return err.EventHasNotErr
	}
	return s.event.DispatchEvent(sourceId, event)
}

type Ev_newActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type Ev_delActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type Ev_newSession struct {
	Host string
}

type Ev_delSession struct {
	Host string
}

type Ev_clusterUpdate struct {
	ActorId string
	Host    string
	Add     bool
}
