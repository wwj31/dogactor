package actor_msg

import (
	"github.com/gogo/protobuf/proto"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"sync"
)

var (
	_msgPool sync.Pool
)

func init() {
	_msgPool.New = func() interface{} { return &ActorMessage{pool: &_msgPool} }
}

func NewActorMessage() *ActorMessage {
	msg := _msgPool.Get().(*ActorMessage)
	return msg
}

type ActorMessage struct {
	pool    *sync.Pool
	payload interface{}

	SourceId   string            `protobuf:"bytes,1,opt,name=SourceId,proto3" json:"SourceId,omitempty"`
	TargetId   string            `protobuf:"bytes,2,opt,name=TargetId,proto3" json:"TargetId,omitempty"`
	RequestId  string            `protobuf:"bytes,3,opt,name=RequestId,proto3" json:"RequestId,omitempty"`
	MsgName    string            `protobuf:"bytes,4,opt,name=MsgName,proto3" json:"MsgName,omitempty"`
	Data       []byte            `protobuf:"bytes,5,opt,name=Data,proto3" json:"Data,omitempty"`
	MapCarrier map[string]string `protobuf:"bytes,6,rep,name=MapCarrier,proto3" json:"MapCarrier,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (a *ActorMessage) Payload() interface{} {
	return a.payload
}

func (a *ActorMessage) Free() {
	if a.pool != nil {
		a.payload = nil
		a.SourceId = ""
		a.TargetId = ""
		a.RequestId = ""
		a.MsgName = ""
		a.MapCarrier = nil
		a.Data = nil
		a.pool.Put(a)
	}
}

func (a *ActorMessage) SetPayload(v interface{}) {
	a.payload = v
}

func (a *ActorMessage) Parse(pi *tools.ProtoIndex) {
	if pi == nil {
		log.SysLog.Errorf("protoIndex is nil")
		return
	}
	pt, ok := pi.FindMsgByName(a.MsgName)
	if !ok {
		log.SysLog.Errorf("msg not found", "MsgName", a.MsgName)
		return
	}

	if a.Data == nil {
		return
	}

	if err := proto.Unmarshal(a.Data, pt.(proto.Message)); err != nil {
		log.SysLog.Errorf("Unmarshal failed", "err", err, "MsgName", a.MsgName)
		return
	}

	a.SetPayload(pt)
}
