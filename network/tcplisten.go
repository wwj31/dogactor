package network

import (
	"net"
	"strconv"
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

func (s *TcpListener) Start(exceptPort ...int) error {
	return s.listen(exceptPort...)
}

func (s *TcpListener) Port() int {
	str := strings.Split(s.listener.Addr().String(), ":")
	port, _ := strconv.Atoi(str[len(str)-1])
	return port
}

func (s *TcpListener) inExceptPort(port int, exceptPort ...int) bool {
	for _, p := range exceptPort {
		if port == p {
			return true
		}
	}
	return false
}

func (s *TcpListener) listen(exceptPort ...int) error {
	for {
		var err error
		s.listener, err = net.Listen("tcp", s.addr)
		if err != nil {
			return err
		}

		str := strings.Split(s.listener.Addr().String(), ":")
		port, _ := strconv.Atoi(str[len(str)-1])
		if s.inExceptPort(port, exceptPort...) {
			log.SysLog.Warnw("listen the port in except the port ",
				"port", port, "exceptPort", exceptPort)
			continue
		}
		break
	}

	go tools.Try(func() {
		for {
			if conn, err := s.listener.Accept(); err != nil {
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
