package tools

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/wwj31/dogactor/log"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type ProtoParser struct {
	typemap  map[int32]protoreflect.MessageType // 此map初始化后，所有session都会并发读取
	msgNames map[string]int32                   // 此map初始化后，所有session都会并发读取
}

func NewProtoParser() *ProtoParser {
	return &ProtoParser{
		typemap:  make(map[int32]protoreflect.MessageType),
		msgNames: make(map[string]int32),
	}
}

//协议enum->message自动解析:
//1、不区分大消息
//2、过滤下划线
func (s *ProtoParser) Init(packageName, msgType string) *ProtoParser {
	enums, err := protoregistry.GlobalTypes.FindEnumByName(protoreflect.FullName(packageName + "." + msgType))
	if err != nil {
		panic(fmt.Errorf("OnInit error packageName=%v msgType=%v", packageName, msgType))
		return s
	}

	lowerNames := map[string]protoreflect.MessageType{}
	protoregistry.GlobalTypes.RangeMessages(func(messageType protoreflect.MessageType) bool {
		if !strings.HasPrefix(string(messageType.Descriptor().FullName()), packageName+".") {
			return true
		}
		name := convertMsgName(string(messageType.Descriptor().Name()))
		if _, ok := lowerNames[name]; ok { //防止重名
			panic(fmt.Sprintf("msg name repeated name=%s", messageType.Descriptor().FullName()))
		}
		lowerNames[name] = messageType
		return true
	})

	values := enums.Descriptor().Values()
	for i := 0; i < values.Len(); i++ {
		msgTypeName := string(values.Get(i).Name())
		if strings.Contains(msgTypeName, "SEGMENT_BEGIN") || strings.Contains(msgTypeName, "SEGMENT_END") {
			continue
		}
		ln := convertMsgName(msgTypeName)
		if tp, ok := lowerNames[ln]; ok {
			s.typemap[int32(values.Get(i).Number())] = tp
			s.msgNames[string(tp.Descriptor().FullName())] = int32(values.Get(i).Number())
		} else {
			panic(fmt.Errorf("msg format error msgTypeName=%v", msgTypeName))
		}
	}
	return s
}

func (s *ProtoParser) UnmarshalPbMsg(msgType int32, data []byte) proto.Message {
	tp, ok := s.typemap[msgType]
	if !ok {
		log.SysLog.Errorw("msg not regist parser", "msgId", msgType, "data", data)
		return nil
	}
	msg := tp.New().Interface().(proto.Message)
	err := proto.Unmarshal(data, msg)
	if err != nil {
		log.SysLog.Errorw("msg parse failed", "msgType", msgType, "data", data)
		return nil
	}
	return msg
}

func (s *ProtoParser) MsgIdToName(msgId int32) (msgName string, ok bool) {
	ptype, has := s.typemap[msgId]
	if !has {
		return
	}
	return string(ptype.Descriptor().FullName()), true
}

func (s *ProtoParser) MsgNameToId(msgName string) (msgId int32, ok bool) {
	id, has := s.msgNames[msgName]
	if !has {
		return
	}
	return id, true
}

func convertMsgName(msgName string) (name string) {
	words := strings.Split(msgName, "_")
	for _, word := range words {
		name += strings.ToLower(word)
	}
	return
}

func MsgName(msg proto.Message) string {
	return string(proto.MessageReflect(msg).Descriptor().FullName())
}

func FindMsgByName(msgName string) (protoreflect.MessageType, error) {
	return protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgName))
}
