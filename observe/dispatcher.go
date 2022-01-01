package observe

import (
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type EventType = int
type (
	Event    interface{ Type() EventType }
	Handle   func(ev interface{})
	Observer interface {
		Dispatch(event Event)
		AddListener(eventType EventType, callback Handle)
		RemoveListener(eventType EventType, callback Handle)
	}

	eventDispatcher struct {
		listeners map[EventType][]Handle
	}
)

func New() Observer {
	return &eventDispatcher{
		listeners: make(map[EventType][]Handle),
	}
}

func (s *eventDispatcher) Dispatch(event Event) {
	list, ok := s.listeners[event.Type()]
	if !ok {
		//log.KV("etype", event.Type()).WarnStack(1, "event has no listener")
		return
	}

	for _, fun := range list {
		tools.Try(func() {
			fun(event)
		})
	}
}

func (s *eventDispatcher) AddListener(eventType EventType, callback Handle) {
	_, ok := s.listeners[eventType]
	if !ok {
		s.listeners[eventType] = make([]Handle, 0, 8)
	}

	list := s.listeners[eventType]
	for i := range list {
		if list[i] == callback {
			log.SysLog.Errorw("AddEventListener repeated", "etype", eventType)
			return
		}
	}
	s.listeners[eventType] = append(list, callback)
}

func (s *eventDispatcher) RemoveListener(eventType EventType, callback Handle) {
	list, ok := s.listeners[eventType]
	if !ok {
		log.SysLog.Errorw("RemoveListenerByType failed", "eventType", eventType)
		return
	}
	for i := range list {
		if list[i] == callback {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	s.listeners[eventType] = list
}
