package conngrpc

import (
	"errors"

	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote/conngrpc/internal"

	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/atomic"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
)

// 管理所有远端的session
type RemoteMgr struct {
	remoteHandler remote.Handler
	listener      *internal.Server

	stop     atomic.Int32
	sessions cmap.ConcurrentMap //host=>session
	clients  cmap.ConcurrentMap //host=>client

	ping   *actor_msg.ActorMessage
	regist *actor_msg.ActorMessage
}

func NewRemoteMgr() *RemoteMgr {
	mgr := &RemoteMgr{
		sessions: cmap.New(),
		clients:  cmap.New(),
	}
	return mgr
}

func (s *RemoteMgr) Start(actorSystem remote.Handler) error {
	s.remoteHandler = actorSystem

	listener, err := internal.NewServer(s.remoteHandler.Address(), func() internal.Handler { return &remoteHandler{remote: s} }, s)
	if err != nil {
		return err
	}

	s.regist = actor_msg.NewActorMessage() // registry message
	s.regist.MsgName = "$regist"
	s.regist.Data = []byte(s.remoteHandler.Address())

	s.listener = listener
	return err
}

func (s *RemoteMgr) Stop() {
	if s.stop.CAS(0, 1) {
		if s.listener != nil {
			s.listener.Stop()
		}
		s.clients.IterCb(func(key string, v interface{}) { v.(*internal.Client).Stop() })
	}
}

func (s *RemoteMgr) NewClient(host string) {
	c := internal.NewClient(host, func() internal.Handler { return &remoteHandler{remote: s, peerHost: host} })
	s.clients.Set(host, c)
	_ = c.Start(true)
}

func (s *RemoteMgr) StopClient(host string) {
	if c, ok := s.clients.Pop(host); ok {
		c.(*internal.Client).Stop()
	}
}

func (s *RemoteMgr) SendMsg(addr string, netMsg *actor_msg.ActorMessage) error {
	session, ok := s.sessions.Get(addr)
	if !ok {
		return errors.New("remote addr not found")
	}
	return session.(*remoteHandler).Send(netMsg)
}

// /////////////////////////////////////// remoteHandler /////////////////////////////////////////////
type remoteHandler struct {
	internal.BaseHandler

	remote   *RemoteMgr
	peerHost string
}

func (s *remoteHandler) OnSessionCreated() {
	_ = s.Send(s.remote.regist)
}

func (s *remoteHandler) OnSessionClosed() {
	if len(s.peerHost) > 0 {
		s.remote.sessions.Remove(s.peerHost)
		s.remote.remoteHandler.OnSessionClosed(s.peerHost)
	}
}

func (s *remoteHandler) OnRecv(msg *actor_msg.ActorMessage) {
	if s.remote == nil {
		s.Stop()
		return
	}

	if msg.MsgName == "$regist" {
		s.peerHost = string(msg.Data)
		s.remote.sessions.Set(s.peerHost, s)
		s.remote.remoteHandler.OnSessionOpened(s.peerHost)
	} else {
		if s.peerHost == "" {
			log.SysLog.Errorw("has not regist", "msg", msg.MsgName)
			return
		}
		s.remote.remoteHandler.OnSessionRecv(msg)
	}
}
