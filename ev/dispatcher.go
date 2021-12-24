package ev

import (
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

/*
	单线程事件机制
*/
type (
	EventType        uint32
	IEvent           interface{ EType() EventType }
	IEventListener   interface{ OnEvent(IEvent) }
	IEventDispatcher interface {
		Dispatch(event IEvent)
		AddEventListener(eventType EventType, listener IEventListener)
		RemoveListenerByEType(eventType EventType, listener IEventListener)
		RemoveListenerAll(listener IEventListener)
	}

	eventDispatcher struct {
		listeners map[EventType][]IEventListener
	}
)

func New() IEventDispatcher {
	return &eventDispatcher{
		listeners: make(map[EventType][]IEventListener),
	}
}

func (s *eventDispatcher) Dispatch(event IEvent) {
	list, ok := s.listeners[event.EType()]
	if !ok {
		//log.KV("etype", event.EType()).WarnStack(1, "event has no listener")
		return
	}

	for _, listener := range list {
		tools.Try(func() {
			listener.OnEvent(event)
		})
	}
}

func (s *eventDispatcher) AddEventListener(eventType EventType, listener IEventListener) {
	_, ok := s.listeners[eventType]
	if !ok {
		s.listeners[eventType] = make([]IEventListener, 0)
	}

	list := s.listeners[eventType]
	for i := range list {
		if list[i] == listener {
			log.SysLog.Errorw("AddEventListener repeated", "etype", eventType)
			return
		}
	}
	s.listeners[eventType] = append(list, listener)
}

func (s *eventDispatcher) RemoveListenerByEType(eventType EventType, listener IEventListener) {
	list, ok := s.listeners[eventType]
	if !ok {
		log.SysLog.Errorw("RemoveListenerByEType failed", "eventType", eventType)
		return
	}
	for i := range list {
		if list[i] == listener {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	s.listeners[eventType] = list
}

func (s *eventDispatcher) RemoveListenerAll(listener IEventListener) {
	for typ := range s.listeners {
		s.RemoveListenerByEType(typ, listener)
	}
}
