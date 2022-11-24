package actor

import (
	"reflect"
	"sync"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/log"
)

type Handler interface{}

type fnMeta struct {
	CallValue reflect.Value
	ArgType   reflect.Type
}

type listener map[string]map[Id]fnMeta // map[evType][actorId]Handler

type evDispatcher struct {
	sync.RWMutex
	sys       *System
	listeners listener
}

func newEvent(s *System) evDispatcher {
	return evDispatcher{listeners: make(listener), sys: s}
}

func (ed *evDispatcher) OnEvent(actorId Id, callback Handler) {
	argType, n := argInfo(callback)
	if n != 1 {
		log.SysLog.Errorw("OnEvent Handler param len !=1",
			"n", n, "actorId", actorId)
		return
	}

	if argType.Kind() == reflect.Ptr {
		log.SysLog.Errorw("OnEvent Handler param must be a value instead of the point",
			"kind", argType.Kind(), "n", n, "actorId", actorId)
		return
	}

	callValue := reflect.ValueOf(callback)

	ed.Lock()
	defer ed.Unlock()

	typ := argType.String()
	if ed.listeners[typ] == nil {
		ed.listeners[typ] = make(map[string]fnMeta)
	}

	ed.listeners[typ][actorId] = fnMeta{
		CallValue: callValue,
		ArgType:   argType,
	}
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

	ed.RLock()
	defer ed.RUnlock()

	if listeners := ed.listeners[rType.String()]; listeners != nil {
		for actorId, fnMeta := range listeners {
			err := ed.sys.Send(sourceId, actorId, "", func() {
				if fnMeta.ArgType.Kind() != rType.Kind() {
					log.SysLog.Errorw("param type kind not equal", "fnMeta.ArgType", fnMeta.ArgType.Kind(), "rType", rType.Kind())
					return
				}
				fnMeta.CallValue.Call([]reflect.Value{reflect.ValueOf(event)})
			})
			if err != nil {
				log.SysLog.Errorw("DispatchEvent send to actor", "actorId", actorId, "err", err)
			}
		}
	}
}

// Dissect the cb Handler's signature
func argInfo(cb Handler) (reflect.Type, int) {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		log.SysLog.Errorw("nats: Handler needs to be a func")
		return nil, 0
	}
	numArgs := cbType.NumIn()
	if numArgs == 0 {
		return nil, numArgs
	}
	return cbType.In(numArgs - 1), numArgs
}
