package actor

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/wwj31/dogactor/actor/err"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
)

type listener map[string]map[string]bool // map[evType][actorId]bool

type evDispatcher struct {
	sync.RWMutex
	sys       *System
	listeners listener
}

func newEvent(s *System) *evDispatcher {
	return &evDispatcher{listeners: make(listener), sys: s}
}

// 注册actor事件
func (ed *evDispatcher) RegistEvent(actorId string, events ...interface{}) error {
	for _, event := range events {
		rtype := reflect.TypeOf(event)
		if rtype.Kind() != reflect.Ptr {
			return fmt.Errorf("%w actorId:%v,event:%v", err.RegisterEventErr, actorId, event)
		}
	}

	ed.Lock()
	defer ed.Unlock()

	for _, event := range events {
		rtype := reflect.TypeOf(event)
		etype := rtype.Elem().Name()
		if ed.listeners[etype] == nil {
			ed.listeners[etype] = make(map[string]bool)
		}
		ed.listeners[etype][actorId] = true
	}
	return nil
}

// 取消actor事件
func (ed *evDispatcher) CancelEvent(actorId string, events ...interface{}) error {
	for _, event := range events {
		rtype := reflect.TypeOf(event)
		if rtype.Kind() != reflect.Ptr {
			return fmt.Errorf(" %w,actorId:%v,event:%v", err.CancelEventErr, actorId, event)
		}
	}
	ed.Lock()
	defer ed.Unlock()

	for _, event := range events {
		rtype := reflect.TypeOf(event)
		etype := rtype.Elem().Name()
		delete(ed.listeners[etype], actorId)
	}
	return nil
}

// 取消actor事件
func (ed *evDispatcher) CancelAll(actorId string) {
	ed.Lock()
	defer ed.Unlock()

	for _, actors := range ed.listeners {
		delete(actors, actorId)
	}
}

// 事件触发
func (ed *evDispatcher) DispatchEvent(sourceId string, event interface{}) error {
	rtype := reflect.TypeOf(event)
	if rtype.Kind() != reflect.Ptr {
		return fmt.Errorf(" %w,actorId:%v,event:%v", err.DispatchEventErr, sourceId, event)
	}

	etype := rtype.Elem().Name()
	wrap := actor_msg.NewEventMessage(event)

	ed.RLock()
	defer ed.RUnlock()

	if listeners := ed.listeners[etype]; listeners != nil {
		for actorId, _ := range listeners {
			ed.sys.Send(sourceId, actorId, "", wrap)
		}
	}
	return nil
}
