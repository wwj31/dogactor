package network

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

func newTcpSession(conn net.Conn, coder ICodec, handler ...INetHandler) *TcpSession {
	session := &TcpSession{
		id:      GenNetSessionId(),
		conn:    conn,
		running: 1,
		coder:   coder,
		handler: handler,
		sendQue: make(chan []byte, 16),
	}
	log.KVs(log.Fields{"sessionId": session.Id(), "local": session.LocalAddr(), "remote": session.RemoteAddr()}).Debug("new tcp session")
	return session
}

type TcpSession struct {
	id      uint32
	conn    net.Conn
	storage sync.Map

	running int32
	coder   ICodec
	handler []INetHandler
	sendQue chan []byte
}

func (s *TcpSession) Type() SessionType    { return TYPE_TCP }
func (s *TcpSession) Id() uint32           { return s.id }
func (s *TcpSession) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *TcpSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }
func (s *TcpSession) RemoteIP() string {
	addr := s.RemoteAddr()
	switch v := addr.(type) {
	case *net.UDPAddr:
		if ip := v.IP.To4(); ip != nil {
			return ip.String()
		}
	case *net.TCPAddr:
		if ip := v.IP.To4(); ip != nil {
			return ip.String()
		}
	case *net.IPAddr:
		if ip := v.IP.To4(); ip != nil {
			return ip.String()
		}
	}
	return ""
}

func (s *TcpSession) StoreKV(key, value interface{}) { s.storage.Store(key, value) }
func (s *TcpSession) DeleteKV(key interface{})       { s.storage.Delete(key) }
func (s *TcpSession) Load(key interface{}) (value interface{}, ok bool) {
	return s.storage.Load(key)
}

func (s *TcpSession) start() {
	for _, v := range s.handler {
		tools.Try(func() {
			v.OnSessionCreated(s)
		}, nil)
	}
	go s.read()
	go s.write()
}

func (s *TcpSession) Stop() {
	if atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		close(s.sendQue)
		err := s.conn.Close()
		if err != nil {
			log.KV("actorerr", err).Error("stop error")
		}
		for _, v := range s.handler {
			tools.Try(v.OnSessionClosed)
		}
		log.KV("sessionId", s.id).Debug("tcp session close")
	}
}

func (s *TcpSession) SendMsg(msg []byte) error {
	if atomic.LoadInt32(&s.running) == 0 {
		return errors.New("tcp session was stop")
	}
	tools.Try(func() { s.sendQue <- msg }, nil)
	return nil
}

func (s *TcpSession) read() {
	buffer := make([]byte, 1024*8)

	for {
		if err := s.conn.SetReadDeadline(tools.Now().Add(time.Second * 30)); err != nil {
			log.KVs(log.Fields{"sessionId": s.Id(), "actorerr": err}).Info("tcp read SetReadDeadline")
			break
		}

		n, err := s.conn.Read(buffer)
		if err != nil {
			if operr, ok := err.(*net.OpError); ok && (operr.Err == syscall.EAGAIN || operr.Err == syscall.EWOULDBLOCK) { //没数据了
				continue
			}
			log.KVs(log.Fields{"sessionId": s.Id(), "actorerr": err}).Info("tcp read buff failed")
			break
		}

		if n == 0 {
			continue
		}

		datas, err := s.coder.Decode(buffer[:n])
		if err != nil {
			log.KVs(log.Fields{"sessionId": s.Id(), "actorerr": err}).Info("tcp decode failed")
			break
		}

		for _, d := range datas {
			for _, v := range s.handler {
				tools.Try(func() { v.OnRecv(d) }, nil)
			}
		}
	}
	s.Stop()
}

func (s *TcpSession) write() {
	for data := range s.sendQue {
		if err := s.conn.SetWriteDeadline(tools.Now().Add(time.Second * 5)); err != nil {
			log.KVs(log.Fields{"sessionId": s.Id(), "actorerr": err}).Info("tcp read SetWriteDeadline")
			break
		}

		if _, err := s.conn.Write(s.coder.Encode(data)); err != nil {
			log.KVs(log.Fields{"sessionId": s.Id(), "actorerr": err}).Info("tcp session write error")
			break
		}
	}
	s.Stop()
}
