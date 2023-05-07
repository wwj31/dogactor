package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/wwj31/dogactor/tools"
)

var ErrRecvLen = errors.New("data is too long")

// the message package base infrastructure
//
//		├──── 4bytes ─────┼────────────── body size ─────────────────┤
//		┏━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
//		┃ head(body size) ┃               body (data)                ┃
//		┗━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
//	    ├──────────── MAX Size:StreamCode.MaxDecode ─────────────────┤

const (
	StreamHeadLen = 4 * tools.B
	MaxMsgLen     = 100 * tools.MB
)

type StreamCode struct {
	MaxDecode uint64
	msgLen    uint64
	buffer    bytes.Buffer
}

func (s *StreamCode) Decode(data []byte) ([][]byte, error) {
	s.buffer.Write(data)

	var ret [][]byte = nil
	for uint64(s.buffer.Len()) >= StreamHeadLen {
		if s.msgLen == 0 {
			d := s.buffer.Bytes()
			s.msgLen = uint64(binary.BigEndian.Uint32(d[:StreamHeadLen]))
			if s.MaxDecode > 0 && s.msgLen > s.MaxDecode && s.msgLen > MaxMsgLen {
				return nil, ErrRecvLen
			}
		}

		if s.msgLen+StreamHeadLen > uint64(s.buffer.Len()) {
			break
		}

		d := make([]byte, s.msgLen+StreamHeadLen)
		n, err := s.buffer.Read(d)
		if uint64(n) != s.msgLen+StreamHeadLen || err != nil {
			s.msgLen = 0
			continue
		}
		s.msgLen = 0
		ret = append(ret, d[StreamHeadLen:])
	}
	return ret, nil
}

func (s *StreamCode) Encode(data []byte) []byte {
	d := make([]byte, StreamHeadLen)
	binary.BigEndian.PutUint32(d, uint32(len(data)))
	data = append(d, data...)
	return data
}
