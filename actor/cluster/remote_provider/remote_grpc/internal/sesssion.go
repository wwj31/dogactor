package internal

import (
	"errors"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"sync"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/log"
	"github.com/wwj31/dogactor/tools"
)

type Stream interface {
	Send(*actor_msg.ActorMessage) error
	Recv() (*actor_msg.ActorMessage, error)
}

type Handler interface {
	OnRecv(*actor_msg.ActorMessage)
	OnSessionCreated()
	OnSessionClosed()
	setSession(*Session)
}

type BaseHandler struct {
	*Session
}

func (h *BaseHandler) setSession(session *Session) { h.Session = session }

func NewSession(stream Stream, handler Handler) *Session {
	session := &Session{
		id:      SessionId(),
		chSend:  make(chan *actor_msg.ActorMessage, 1024),
		handler: handler,
		stream:  stream,
	}
	handler.setSession(session)
	return session
}

type Session struct {
	ext     sync.Map
	id      int64
	running atomic.Int32
	chSend  chan *actor_msg.ActorMessage

	handler Handler
	stream  Stream
}

func (s *Session) Id() int64 { return s.id }

func (s *Session) Stop() {
	if s.running.CAS(0, 1) {
		close(s.chSend)
		log.SysLog.Debugw("session closed","id", s.id)
	}
}

func (s *Session) Send(msg *actor_msg.ActorMessage) (err error) {
	if s.running.Load() != 0 {
		return errors.New("session has closed")
	}
	tools.Try(func() { s.chSend <- msg }, func(ex interface{}) { err = errors.New("session has closed") })
	return err
}

func (s *Session) start(block bool) {
	wait := sync.WaitGroup{}
	wait.Add(2)

	tools.Try(s.handler.OnSessionCreated)
	go func() { s.read(); wait.Done() }()
	go func() { s.write(); wait.Done() }()

	if block {
		wait.Wait()
	}
}

func (s *Session) read() {
	for {
		msg, err := s.stream.Recv()
		if err != nil {
			log.SysLog.Debugw("grpc recv error","error", err,"id", s.id)
			break
		}
		tools.Try(func() {
			s.handler.OnRecv(msg)
		})
	}
	tools.Try(s.handler.OnSessionClosed)
	s.Stop()
}

func (s *Session) write() {
	for msg := range s.chSend {
		if err := s.stream.Send(msg); err != nil {
			log.SysLog.Debugw("send msg error","error", err,"id", s.id)
			break
		}
		msg.Free()
	}
	s.Stop()
	//必须与stream.Send同线程
	if s, ok := s.stream.(grpc.ClientStream); ok {
		s.CloseSend()
	}
}

var SessionId = _gen_session_id()

func _gen_session_id() func() int64 {
	_session_gen_id := atomic.NewInt64(tools.Now().UnixNano())
	return func() int64 { return _session_gen_id.Add(1) }
}
