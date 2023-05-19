package network

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

func newWSSession(conn *websocket.Conn, coder DecodeEncoder, handler ...SessionHandler) *WSSession {
	ctx, cancel := context.WithCancel(context.Background())
	session := &WSSession{
		id:      GenSessionId(),
		conn:    conn,
		ctx:     ctx,
		cancel:  cancel,
		coder:   coder,
		handler: handler,
		sendQue: make(chan []byte, 16),
	}
	session.running.Store(1)

	log.SysLog.Debugw("new ws session",
		"sessionId", session.Id(),
		"local", session.LocalAddr(),
		"remote", session.RemoteAddr())

	return session
}

type WSSession struct {
	id      uint64
	conn    *websocket.Conn
	ctx     context.Context
	cancel  context.CancelFunc
	storage sync.Map

	running     atomic.Value // 1.running 0.stop
	coder       DecodeEncoder
	handler     []SessionHandler
	sendQue     chan []byte
	messageType int
}

func (s *WSSession) Type() SessionType    { return TypeWS }
func (s *WSSession) Id() uint64           { return s.id }
func (s *WSSession) LocalAddr() net.Addr  { return s.conn.LocalAddr() }
func (s *WSSession) RemoteAddr() net.Addr { return s.conn.RemoteAddr() }
func (s *WSSession) RemoteIP() string {
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

func (s *WSSession) Context() context.Context       { return s.ctx }
func (s *WSSession) StoreKV(key, value interface{}) { s.storage.Store(key, value) }
func (s *WSSession) DeleteKV(key interface{})       { s.storage.Delete(key) }
func (s *WSSession) Load(key interface{}) (value interface{}, ok bool) {
	return s.storage.Load(key)
}

func (s *WSSession) start() context.Context {
	for _, v := range s.handler {
		tools.Try(func() {
			v.OnSessionCreated(s)
		}, nil)
	}

	go s.read()
	go s.write()
	return s.ctx
}

func (s *WSSession) Stop() {
	if s.running.CompareAndSwap(1, 0) {
		close(s.sendQue)
		err := s.conn.Close()
		if err != nil {
			log.SysLog.Errorw("tcp session stop error", "err", err)
		}

		for _, v := range s.handler {
			tools.Try(v.OnSessionClosed)
		}

		s.cancel()
		log.SysLog.Infow("tcp session close", "sessionId", s.id)
	}
}

func (s *WSSession) SendMsg(msg []byte) error {
	if s.running.Load() == 0 {
		return errors.New("tcp session was stop")
	}
	tools.Try(func() { s.sendQue <- msg }, nil)
	return nil
}

func (s *WSSession) read() {
	defer s.Stop()

	for {
		mt, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.SysLog.Infof("the client disconnected ", "sessionId", s.id)
			} else {
				log.SysLog.Warnw("ws read buff failed", "sessionId", s.Id(), "err", err)
			}
			break
		}
		s.messageType = mt

		datas, decErr := s.coder.Decode(message)
		if decErr != nil {
			log.SysLog.Errorw("tcp decode failed", "sessionId", s.Id(), "err", decErr)
			break
		}

		for _, d := range datas {
			for _, v := range s.handler {
				tools.Try(func() { v.OnRecv(d) })
			}
		}
	}
}

func (s *WSSession) write() {
	defer s.Stop()

	for data := range s.sendQue {
		if err := s.conn.SetWriteDeadline(time.Now().Add(time.Second * 5)); err != nil {
			log.SysLog.Errorw("tcp read SetWriteDeadline", "sessionId", s.Id(), "err", err)
			break
		}

		if err := s.conn.WriteMessage(s.messageType, s.coder.Encode(data)); err != nil {
			log.SysLog.Warnw("tcp session write error", "sessionId", s.Id(), "err", err)
			break
		}
	}
}
