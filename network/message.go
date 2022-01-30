package network

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type Message struct {
	msgId int32
	pb    proto.Message
	bytes []byte
}

func (s *Message) Buffer() []byte {
	if len(s.bytes) > 0 {
		return s.bytes
	}
	if s.pb != nil {
		data, err := proto.Marshal(s.pb)
		if err != nil {
			log.SysLog.Errorw("marshal pb failed", "err", err)
			return nil
		}
		return append(Uint32ToByte4(uint32(s.msgId)), data...)
	}

	return Uint32ToByte4(uint32(s.msgId))
}

func (s *Message) MsgId() int32         { return s.msgId }
func (s *Message) Proto() proto.Message { return s.pb }

func (s *Message) parse(data []byte, mm *tools.ProtoIndex) *Message {
	dlen := len(data)
	if dlen < 4 {
		log.SysLog.Errorw("actorerr msg length", "data", data)
		return nil
	}

	s.msgId = int32(Byte4ToUint32(data[:4]))
	pb := mm.UnmarshalPbMsg(s.msgId, data[4:])
	if nil == pb {
		return nil
	}

	s.pb = pb
	s.bytes = data
	return s
}

// 业务逻辑层主要使用接口
func NewPbMessage(pb proto.Message, msgId int32) *Message {
	msg := &Message{msgId: msgId, pb: pb}
	return msg
}

func CombineMsgWithId(msgId int32, data []byte) []byte {
	return append(Uint32ToByte4(uint32(msgId)), data...)
}

// 远端actor通信主要使用接口
func NewBytesMessageParse(data []byte, mm *tools.ProtoIndex) *Message {
	msg := &Message{}
	return msg.parse(data, mm)
}

func Byte4ToUint32(data []byte) (result uint32) {
	buff := bytes.NewBuffer(data)
	binary.Read(buff, binary.BigEndian, &result)
	return
}

func Uint32ToByte4(data uint32) (result []byte) {
	buff := bytes.NewBuffer([]byte{})
	binary.Write(buff, binary.BigEndian, data)
	return buff.Bytes()
}
