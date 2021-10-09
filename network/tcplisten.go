package network

import (
	"net"
	"strings"
	"sync/atomic"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type OptionListen func(l *TcpListener)

func StartTcpListen(addr string, newCodec func() ICodec, newHanlder func() INetHandler, op ...OptionListen) INetListener {
	l := &TcpListener{
		addr:       addr,
		newCodec:   newCodec,
		newHanlder: newHanlder,
	}

	for _, f := range op {
		f(l)
	}
	return l
}

type TcpListener struct {
	addr       string
	listener   net.Listener
	running    int32
	newCodec   func() ICodec
	newHanlder func() INetHandler
}

func (s *TcpListener) Start() error {
	err := s.listen()
	if err == nil {
		log.KV("addr", s.addr).Info("tcp listen success")
		return nil
	}
	log.KV("addr", s.addr).KV("actorerr", err).Error("tcp listen failed")
	return err
}

func (s *TcpListener) listen() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener

	tools.GoEngine(func() {
		for {
			if conn, err := listener.Accept(); err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}
				log.KV("addr", s.addr).KV("actorerr", err).Warn("tcp accept error")
			} else {
				newTcpSession(conn, s.newCodec(), s.newHanlder()).start()
			}
		}
		s.Stop()
	})
	return nil
}

func (s *TcpListener) Stop() {
	if atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		log.KV("addr", s.addr).Info("tcp stop listen")
		if l := s.listener; l != nil {
			_ = s.listener.Close()
		}
	}
}
