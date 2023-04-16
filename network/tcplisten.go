package network

import (
	"net"
	"strings"
	"sync/atomic"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

func StartTcpListen(addr string, newCodec func() DecodeEncoder, newHandler func() SessionHandler) Listener {
	l := &TcpListener{
		addr:       addr,
		newCodec:   newCodec,
		newHandler: newHandler,
	}
	return l
}

type TcpListener struct {
	addr       string
	listener   net.Listener
	running    int32
	newCodec   func() DecodeEncoder
	newHandler func() SessionHandler
}

func (s *TcpListener) Start() error {
	return s.listen()
}

func (s *TcpListener) listen() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener

	go tools.Try(func() {
		for {
			if conn, err := listener.Accept(); err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}
				log.SysLog.Errorw("tcp accept error", "addr", s.addr, "err", err)
			} else {
				newTcpSession(conn, s.newCodec(), s.newHandler()).start()
			}
		}
		s.Stop()
	})
	return nil
}

func (s *TcpListener) Stop() {
	if atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		log.SysLog.Infof("tcp stop listen", "addr", s.addr)
		if l := s.listener; l != nil {
			_ = s.listener.Close()
		}
	}
}
