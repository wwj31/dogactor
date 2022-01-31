package actor_msg

import (
	"github.com/gogo/protobuf/proto"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
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

type Message interface {
	Free()
	GetSourceId() string
	GetTargetId() string
	GetRequestId() string
	GetMsgName() string
	Message() interface{}
	String() string
}

func NewActorMessage() *ActorMessage {
	msg := _msgPool.Get().(*ActorMessage)
	atomic.StoreInt32(&msg.free, 1)
	return msg
}

type ActorMessage struct {
	pool    *sync.Pool
	free    int32
	message interface{}

	SourceId  string `protobuf:"bytes,1,opt,name=SourceId,proto3" json:"SourceId,omitempty"`
	TargetId  string `protobuf:"bytes,2,opt,name=TargetId,proto3" json:"TargetId,omitempty"`
	RequestId string `protobuf:"bytes,3,opt,name=RequestId,proto3" json:"RequestId,omitempty"`
	MsgName   string `protobuf:"bytes,4,opt,name=MsgName,proto3" json:"MsgName,omitempty"`
	Data      []byte `protobuf:"bytes,5,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (s *ActorMessage) Message() interface{} {
	return s.message
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

func (msg *ActorMessage) LockFree() {
	atomic.StoreInt32(&msg.free, 0)
}
func (msg *ActorMessage) UnlockFree() {
	atomic.StoreInt32(&msg.free, 1)
}

func (s *ActorMessage) SetMessage(v interface{}) {
	s.message = v
}

func (msg *ActorMessage) Fill(pi *tools.ProtoIndex) interface{} {
	if pi == nil {
		log.SysLog.Errorf("protoIndex is nil")
		return nil
	}
	pt ,ok := pi.FindMsgByName(msg.MsgName)
	if !ok{
		log.SysLog.Errorf("msg not found",  "MsgName", msg.MsgName)
		return nil
	}

	if err := proto.Unmarshal(msg.Data, pt.(proto.Message)); err != nil {
		log.SysLog.Errorf("Unmarshal failed", "err", err, "MsgName", msg.MsgName)
		return nil
	}
	return pt
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
