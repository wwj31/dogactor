package remote_tcp

import (
	"errors"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote_provider"
	"github.com/wwj31/dogactor/expect"
	"go.uber.org/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/network"
)

type RemoteMgr struct {
	remoteHandler remote_provider.RemoteHandler
	listener      network.Listener

	stop     atomic.Int32
	sessions cmap.ConcurrentMap //host=>session
	clients  cmap.ConcurrentMap //host=>client

	ping     []byte
	registry []byte
}

func NewRemoteMgr() *RemoteMgr {
	mgr := &RemoteMgr{
		sessions: cmap.New(),
		clients:  cmap.New(),
	}
	return mgr
}

func (s *RemoteMgr) Start(h remote_provider.RemoteHandler) error {
	s.remoteHandler = h

	listener := network.StartTcpListen(s.remoteHandler.Address(),
		func() network.DecodeEncoder { return &network.StreamCode{} },
		func() network.NetSessionHandler { return &remoteHandler{remote: s} },
	)

	err := listener.Start()
	if err != nil {
		return err
	}

	ping := actor_msg.NewActorMessage() // ping message
	ping.MsgName = "$ping"
	s.ping, err = ping.Marshal()
	if err != nil {
		return err
	}
	ping.Free()

	reg := actor_msg.NewActorMessage() // registry message
	reg.MsgName = "$registry"
	reg.Data = []byte(s.remoteHandler.Address())
	s.registry, err = reg.Marshal()
	if err != nil {
		return err
	}
	reg.Free()

	s.listener = listener
	go s.keepAlive()

	return err
}

func (s *RemoteMgr) Stop() {
	if s.stop.CAS(0, 1) {
		if s.listener != nil {
			s.listener.Stop()
		}
		s.clients.IterCb(func(key string, v interface{}) { v.(network.Client).Stop() })
	}
}

func (s *RemoteMgr) NewClient(host string) {
	c := network.NewTcpClient(host, func() network.DecodeEncoder { return &network.StreamCode{} })
	c.AddHandler(func() network.NetSessionHandler { return &remoteHandler{remote: s, peerHost: host} })
	s.clients.Set(host, c)
	expect.Nil(c.Start(true))
}

func (s *RemoteMgr) StopClient(host string) {
	if c, ok := s.clients.Pop(host); ok {
		c.(network.Client).Stop()
	}
}

func (s *RemoteMgr) keepAlive() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		if s.stop.Load() == 1 {
			break
		}
		select {
		case <-ticker.C:
			s.sessions.IterCb(func(key string, v interface{}) {
				remote, _ := v.(*remoteHandler)
				err := v.(*remoteHandler).SendMsg(s.ping)
				if err != nil {
					log.SysLog.Errorw("keepAlive SendMsg",
						"err", err,
						"remote", remote.RemoteIP())
				}
			})
		}
	}
}

func (s *RemoteMgr) SendMsg(addr string, bytes []byte) error {
	session, ok := s.sessions.Get(addr)
	if !ok {
		return errors.New("remote addr not found")
	}

	return session.(*remoteHandler).SendMsg(bytes)
}

///////////////////////////////////////// remoteHandler /////////////////////////////////////////////
type remoteHandler struct {
	network.NetSession
	remote   *RemoteMgr
	peerHost string
}

func (s *remoteHandler) OnSessionCreated(sess network.NetSession) {
	s.NetSession = sess
	err := sess.SendMsg(s.remote.registry)
	if err != nil {
		log.SysLog.Errorw("OnSessionCreated error", "err", err)
	}
}

func (s *remoteHandler) OnSessionClosed() {
	if len(s.peerHost) > 0 {
		s.remote.sessions.Remove(s.peerHost)
		s.remote.remoteHandler.OnSessionClosed(s.peerHost)
	}
}

func (s *remoteHandler) OnRecv(data []byte) {
	if s.remote == nil {
		s.Stop()
		return
	}

	msg := actor_msg.NewActorMessage() // remote recv message
	err := msg.Unmarshal(data)
	if err != nil {
		log.SysLog.Errorf("unmarshal msg failed", "err", err)
		return
	}

	switch msg.MsgName {
	case "$registry":
		s.peerHost = string(msg.Data)
		s.remote.sessions.Set(s.peerHost, s)
		s.remote.remoteHandler.OnSessionOpened(s.peerHost)
	case "$ping":

	default:
		if s.peerHost == "" {
			log.SysLog.Errorf("has not registry", "msg", msg.MsgName)
			return
		}

		s.remote.remoteHandler.OnSessionRecv(msg)
	}
}
