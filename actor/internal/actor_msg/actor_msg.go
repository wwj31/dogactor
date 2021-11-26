package actor_msg

import (
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	_msgPool   sync.Pool
	_eventPool sync.Pool
)

func init() {
	_msgPool.New = func() interface{} { return &ActorMessage{pool: &_msgPool} }
	_eventPool.New = func() interface{} { return &EventMessage{pool: &_eventPool} }
}

// actor邮箱消息基类
type IMessage interface {
	Free()
	String() string
}

func NewLocalActorMessage(sourceId, targetId, requestId string, message interface{}) *ActorMessage {
	msg := _msgPool.Get().(*ActorMessage)
	atomic.StoreInt32(&msg.free, 1)

	msg.SourceId = sourceId
	msg.TargetId = targetId
	msg.RequestId = requestId
	msg.message = message
	return msg
}

func NewNetActorMessage(sourceId, targetId, requestId string, msgName string, data []byte) *ActorMessage {
	msg := _msgPool.Get().(*ActorMessage)
	atomic.StoreInt32(&msg.free, 1)

	msg.SourceId = sourceId
	msg.TargetId = targetId
	msg.MsgName = msgName
	msg.Data = data
	msg.RequestId = requestId
	return msg
}

type ActorMessage struct {
	pool    *sync.Pool
	free    int32
	message interface{}

	SourceId             string   `protobuf:"bytes,1,opt,name=SourceId,json=sourceId,proto3" json:"SourceId,omitempty"`
	TargetId             string   `protobuf:"bytes,2,opt,name=TargetId,json=targetId,proto3" json:"TargetId,omitempty"`
	RequestId            string   `protobuf:"bytes,3,opt,name=RequestId,json=requestId,proto3" json:"RequestId,omitempty"`
	MsgName              string   `protobuf:"bytes,4,opt,name=MsgName,json=msgName,proto3" json:"MsgName,omitempty"`
	Data                 []byte   `protobuf:"bytes,5,opt,name=Data,json=data,proto3" json:"Data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (msg *ActorMessage) Free() {
	if msg.pool != nil && atomic.CompareAndSwapInt32(&msg.free, 1, 0) {
		msg.message = nil
		msg.SourceId, msg.TargetId = "", ""
		msg.MsgName = ""
		msg.Data = nil
		msg.pool.Put(msg)
	}
}

func (s *ActorMessage) Message() interface{} {
	return s.message
}

func NewEventMessage(actEvent interface{}) *EventMessage {
	event := _eventPool.Get().(*EventMessage)
	atomic.StoreInt32(&event.free, 1)
	event.actEvent = actEvent
	return event
}

type EventMessage struct {
	pool *sync.Pool
	free int32

	actEvent interface{}
}

func (s *EventMessage) ActEvent() interface{} {
	return s.actEvent
}

func (s *EventMessage) Free() {
	if s.pool != nil && atomic.CompareAndSwapInt32(&s.free, 1, 0) {
		s.actEvent = nil
		s.pool.Put(s)
	}
}

func (s *EventMessage) String() string {
	return reflect.TypeOf(s.actEvent).Elem().Name()
}
