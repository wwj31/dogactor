package tools

import (
	"github.com/golang/protobuf/proto"
	"github.com/wwj31/dogactor/log"
	"reflect"
	"strings"
)

type GetProtoByName func(name string) (interface{},bool)
type EnumIdx struct {
	PackageName string
	Enum2Name  map[int32]string
	name2Enum map[string]int32
}

type ProtoIndex struct {
	structByName GetProtoByName
	enum EnumIdx
}

func NewProtoIndex(f GetProtoByName,enum EnumIdx) *ProtoIndex {
	pi :=  &ProtoIndex{
		structByName: f,
		enum: enum,
	}
	newName2Enum := make(map[string]int32,len(pi.enum.Enum2Name))
	for e,name := range pi.enum.Enum2Name{
		lowerName := convertMsgName(name)
		newName := pi.enum.PackageName + "." + lowerName
		pi.enum.Enum2Name[e] = newName
		newName2Enum[newName] = e
	}
	pi.enum.name2Enum = newName2Enum
	return pi
}

func (s *ProtoIndex) UnmarshalPbMsg(msgId int32, data []byte) proto.Message {
	ptName, ok := s.enum.Enum2Name[msgId]
	if !ok {
		log.SysLog.Errorw("msg not regist parser", "msgId", msgId, "data", data)
		return nil
	}

	v,ok := s.structByName(ptName)
	if !ok{
		log.SysLog.Errorw("not found ptName", "msgId", msgId, "ptName", ptName)
		return nil
	}

	msg := v.(proto.Message)
	err := proto.Unmarshal(data, msg)
	if err != nil {
		log.SysLog.Errorw("msg parse failed", "msgId", msgId, "data", data)
		return nil
	}
	return msg
}

func (s *ProtoIndex) MsgIdToName(msgId int32) (msgName string, ok bool) {
	name,ok := s.enum.Enum2Name[msgId]
	if !ok {
		return
	}

	return name, true
}

func (s *ProtoIndex) MsgNameToId(msgName string) (msgId int32, ok bool) {
	id, ok := s.enum.name2Enum[msgName]
	if !ok {
		return
	}
	return id, true
}

func convertMsgName(msgName string) (name string) {
	words := strings.Split(msgName, "_")
	for _, word := range words {
		lower := strings.ToLower(word)
		if len(words) > 0{
			runes := []rune(lower)
			if 97 <= runes[0] && runes[0] <= 122{
				runes[0] -= 32
			}
			lower = string(runes)
		}
		name += lower
	}
	return
}

func (s *ProtoIndex)MsgName(msg proto.Message) string {
	name := reflect.TypeOf(msg).String()
	if len(name) > 0 && name[0] == '*'{
		name = name[1:]
	}
	return name
}

func (s *ProtoIndex)FindMsgByName(msgName string) (proto.Message, bool) {
	v,ok := s.structByName(msgName)
	if !ok{
		return nil,false
	}
	return v.(proto.Message),true
}
