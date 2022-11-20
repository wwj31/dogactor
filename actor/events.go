package actor

import (
	"fmt"
	"github.com/wwj31/dogactor/log"
	"reflect"
	"sync"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
)

type listener map[string]map[Id]struct{} // map[evType][actorId]bool

type evDispatcher struct {
	sync.RWMutex
	sys       *System
	listeners listener
}

func newEvent(s *System) evDispatcher {
	return evDispatcher{listeners: make(listener), sys: s}
}

func (ed *evDispatcher) RegistEvent(actorId Id, events ...interface{}) error {
	for _, event := range events {
		rtype := reflect.TypeOf(event)
		if rtype.Kind() == reflect.Ptr {
			return fmt.Errorf("%w actorId:%v,event:%v", actorerr.RegisterEventErr, actorId, event)
		}
	}

	ed.Lock()
	defer ed.Unlock()

	for _, event := range events {
		typ := reflect.TypeOf(event).String()
		if ed.listeners[typ] == nil {
			ed.listeners[typ] = make(map[string]struct{})
		}
		ed.listeners[typ][actorId] = struct{}{}
	}
	return nil
}

func (ed *evDispatcher) CancelEvent(actorId Id, events ...interface{}) error {
	for _, event := range events {
		rtype := reflect.TypeOf(event)
		if rtype.Kind() == reflect.Ptr {
			return fmt.Errorf(" %w,actorId:%v,event:%v", actorerr.CancelEventErr, actorId, event)
		}
	}
	ed.Lock()
	defer ed.Unlock()

	for _, event := range events {
		rtype := reflect.TypeOf(event).String()
		delete(ed.listeners[rtype], actorId)
	}
	return nil
}

func (ed *evDispatcher) CancelAll(actorId Id) {
	ed.Lock()
	defer ed.Unlock()

	for _, actors := range ed.listeners {
		delete(actors, actorId)
	}
}

func (ed *evDispatcher) DispatchEvent(sourceId Id, event interface{}) {
	rtype := reflect.TypeOf(event)
	if rtype.Kind() == reflect.Ptr {
		log.SysLog.Errorw("dispatch event type of event is ptr",
			"err", actorerr.DispatchEventErr,
			"actorId", sourceId,
			"event", event,
		)
		return
	}

	etype := rtype.String()
	wrap := actor_msg.NewEventMessage(event)

	ed.RLock()
	defer ed.RUnlock()

	if listeners := ed.listeners[etype]; listeners != nil {
		for actorId, _ := range listeners {
			err := ed.sys.Send(sourceId, actorId, "", wrap)
			if err != nil {
				log.SysLog.Errorw("DispatchEvent send to actor", "actorId", actorId, "err", err)
			}
		}
	}
}
