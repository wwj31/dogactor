package actor

import (
	"github.com/wwj31/dogactor/log"
	"reflect"
	"sync"

	"github.com/wwj31/dogactor/actor/actorerr"
)

type listener map[string]map[Id]func(event interface{}) // map[evType][actorId]bool

type evDispatcher struct {
	sync.RWMutex
	sys       *System
	listeners listener
}

func newEvent(s *System) evDispatcher {
	return evDispatcher{listeners: make(listener), sys: s}
}

func (ed *evDispatcher) OnEvent(actorId Id, event interface{}, callback func(event interface{})) {
	rType := reflect.TypeOf(event)
	if rType.Kind() == reflect.Ptr {
		log.SysLog.Errorw("OnEvent failed", "err", actorerr.RegisterEventErr, "actorId", actorId, "event", event)
		return
	}

	ed.Lock()
	defer ed.Unlock()

	typ := reflect.TypeOf(event).String()
	if ed.listeners[typ] == nil {
		ed.listeners[typ] = make(map[string]func(interface{}))
	}
	ed.listeners[typ][actorId] = callback
}

func (ed *evDispatcher) CancelEvent(actorId Id, events ...interface{}) {
	for _, event := range events {
		rtype := reflect.TypeOf(event)
		if rtype.Kind() == reflect.Ptr {
			log.SysLog.Errorw("CancelEvent failed", "err", actorerr.CancelEventErr, "actorId", actorId, "event", event)
			return
		}
	}

	ed.Lock()
	defer ed.Unlock()

	for _, event := range events {
		rType := reflect.TypeOf(event).String()
		delete(ed.listeners[rType], actorId)
	}
}

func (ed *evDispatcher) CancelAll(actorId Id) {
	ed.Lock()
	defer ed.Unlock()

	for _, actors := range ed.listeners {
		delete(actors, actorId)
	}
}

func (ed *evDispatcher) DispatchEvent(sourceId Id, event interface{}) {
	rType := reflect.TypeOf(event)
	if rType.Kind() == reflect.Ptr {
		log.SysLog.Errorw("dispatch event type of event is ptr",
			"err", actorerr.DispatchEventErr,
			"actorId", sourceId,
			"event", event,
		)
		return
	}

	ed.RLock()
	defer ed.RUnlock()

	if listeners := ed.listeners[rType.String()]; listeners != nil {
		for actorId, callback := range listeners {
			err := ed.sys.Send(sourceId, actorId, "", func() {
				callback(event)
			})
			if err != nil {
				log.SysLog.Errorw("DispatchEvent send to actor", "actorId", actorId, "err", err)
			}
		}
	}
}
