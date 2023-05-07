package network

import (
	"bytes"
)

type WSCode struct {
	MaxDecode uint64
	msgLen    uint64
	buffer    bytes.Buffer
}

func (s *WSCode) Decode(data []byte) ([][]byte, error) {
	return [][]byte{data}, nil
}

func (s *WSCode) Encode(data []byte) []byte {
	return data
}
