package network

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type DecodeEncoder interface {
	Decode([]byte) ([][]byte, error)
	Encode(data []byte) []byte
}

var ErrRecvLen = errors.New("data is too long")

type EchoCode struct {
	MaxDecode int

	context bytes.Buffer
}

func (s *EchoCode) Decode(data []byte) ([][]byte, error) {
	s.context.Write(data)
	var ret [][]byte = nil
	for {
		d, err := s.context.ReadBytes('\n')
		if err != nil {
			break
		}
		if s.MaxDecode > 0 && len(ret) > s.MaxDecode {
			return nil, ErrRecvLen
		}
		ret = append(ret, d[:len(d)-1])
	}
	return ret, nil
}

func (s *EchoCode) Encode(data []byte) []byte {
	data = append(data, '\n')
	return data
}

/*

	├──── 4bytes ─────┼────────────── body size ─────────────────┤
	┏━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
	┃ head(body size) ┃               body (data)                ┃
	┗━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
    ├──────────── MAX Size:StreamCode.MaxDecode ────────────────┤
*/
type StreamCode struct {
	MaxDecode int

	msglen  uint32
	context bytes.Buffer
}

const STREAM_HEADLEN = 4

func (s *StreamCode) Decode(data []byte) ([][]byte, error) {
	s.context.Write(data)

	var ret [][]byte = nil
	for s.context.Len() >= STREAM_HEADLEN {
		if s.msglen == 0 {
			d := s.context.Bytes()
			s.msglen = binary.BigEndian.Uint32(d[:STREAM_HEADLEN])
			if s.MaxDecode > 0 && int(s.msglen) > s.MaxDecode {
				return nil, ErrRecvLen
			}
		}

		if int(s.msglen)+STREAM_HEADLEN > s.context.Len() {
			break
		}

		d := make([]byte, s.msglen+STREAM_HEADLEN)
		n, err := s.context.Read(d)
		if n != int(s.msglen)+STREAM_HEADLEN || err != nil {
			s.msglen = 0
			continue
		}
		s.msglen = 0
		ret = append(ret, d[STREAM_HEADLEN:])
	}
	return ret, nil
}

func (s *StreamCode) Encode(data []byte) []byte {
	d := make([]byte, STREAM_HEADLEN)
	binary.BigEndian.PutUint32(d, uint32(len(data)))
	data = append(d, data...)
	return data
}
