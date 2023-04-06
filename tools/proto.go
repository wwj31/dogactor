package tools

import (
	"reflect"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/wwj31/dogactor/log"
)

type GetProtoByName func(name string) (interface{}, bool)
type EnumIdx struct {
	PackageName string
	Prefix      string
	Enum2Name   map[int32]string
	Name2Enum   map[string]int32
}

type ProtoIndex struct {
	structByName GetProtoByName
	enum         EnumIdx
}

func NewProtoIndex(f GetProtoByName, enum EnumIdx) *ProtoIndex {
	pi := &ProtoIndex{
		structByName: f,
		enum:         enum,
	}
	return pi
}

func (s *ProtoIndex) UnmarshalPbMsg(msgId int32, data []byte) proto.Message {
	enumName, ok := s.enum.Enum2Name[msgId]
	if !ok {
		log.SysLog.Errorw("the message not registry to the parser", "msgId", msgId, "data", data)
		return nil
	}

	if !strings.HasPrefix(enumName, s.enum.Prefix) {
		log.SysLog.Errorw("the message has not prefix", "msgId", msgId, "enumName", enumName, "prefix", s.enum.Prefix)
		return nil
	}

	_, ptName, _ := strings.Cut(enumName, s.enum.Prefix)
	v, ok := s.structByName(ptName)
	if !ok {
		log.SysLog.Errorw("cannot find enumName", "msgId", msgId, "enumName", enumName)
		return nil
	}

	msg := v.(proto.Message)
	err := proto.Unmarshal(data, msg)
	if err != nil {
		log.SysLog.Errorw("the message parse failed", "msgId", msgId, "data", data)
		return nil
	}
	return msg
}

func (s *ProtoIndex) MsgIdToName(msgId int32) (msgName string, ok bool) {
	enumName, ok := s.enum.Enum2Name[msgId]
	if !ok {
		return
	}
	_, ptName, _ := strings.Cut(enumName, s.enum.Prefix)

	return ptName, ok
}

func (s *ProtoIndex) MsgNameToId(msgName string) (msgId int32, ok bool) {
	enumName := s.enum.Prefix + msgName
	id, ok := s.enum.Name2Enum[enumName]
	if !ok {
		return
	}
	return id, true
}

func (s *ProtoIndex) MsgName(msg proto.Message) string {
	name := reflect.TypeOf(msg).String()
	if len(name) > 0 && name[0] == '*' {
		name = name[1:]
	}
	return name
}

func (s *ProtoIndex) FindMsgByName(msgName string) (proto.Message, bool) {
	v, ok := s.structByName(msgName)
	if !ok {
		return nil, false
	}
	return v.(proto.Message), true
}
