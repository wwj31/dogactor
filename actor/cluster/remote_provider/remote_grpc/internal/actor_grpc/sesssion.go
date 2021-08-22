package actor_grpc

import (
	"errors"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"sync"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type IStream interface {
	Send(*actor_msg.ActorMessage) error
	Recv() (*actor_msg.ActorMessage, error)
}

type IHandler interface {
	OnRecv(*actor_msg.ActorMessage)
	OnSessionCreated()
	OnSessionClosed()
	setSession(*Session)
}

type BaseHandler struct {
	*Session
}

func (h *BaseHandler) setSession(session *Session) { h.Session = session }

func NewSession(stream IStream, handler IHandler) *Session {
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

	handler IHandler
	stream  IStream
}

func (s *Session) Id() int64 { return s.id }

func (s *Session) Stop() {
	if s.running.CAS(0, 1) {
		close(s.chSend)
		log.KV("id", s.id).Debug("session closed")
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

	tools.Try(s.handler.OnSessionCreated, nil)
	tools.GoEngine(func() { s.read(); wait.Done() })
	tools.GoEngine(func() { s.write(); wait.Done() })

	if block {
		wait.Wait()
	}
}

func (s *Session) read() {
	for {
		msg, err := s.stream.Recv()
		if err != nil {
			log.KV("id", s.id).KV("error", err).Warn("grpc recv error")
			break
		}
		tools.Try(func() { s.handler.OnRecv(msg) }, nil)
	}
	tools.Try(s.handler.OnSessionClosed, nil)
	s.Stop()
}

func (s *Session) write() {
	for msg := range s.chSend {
		if err := s.stream.Send(msg); err != nil {
			log.KV("id", s.id).KV("error", err).Error("send msg error")
			break
		}
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
