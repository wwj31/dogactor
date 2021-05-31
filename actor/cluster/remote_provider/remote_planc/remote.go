package remote_planc

import (
	"errors"
	gproto "github.com/golang/protobuf/proto"
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/atomic"
	"time"

	"github.com/wwj31/godactor/actor"
	"github.com/wwj31/godactor/actor/internal/actor_msg"
	"github.com/wwj31/godactor/log"
	"github.com/wwj31/godactor/network"
	"github.com/wwj31/godactor/tools"
)

// 管理所有远端的session
type RemoteMgr struct {
	actorSystem *actor.ActorSystem
	listener    network.INetListener

	stop     atomic.Int32
	sessions cmap.ConcurrentMap //host=>session
	clients  cmap.ConcurrentMap //host=>client

	ping   []byte
	regist []byte
}

func NewRemoteMgr() *RemoteMgr {
	mgr := &RemoteMgr{
		sessions: cmap.New(),
		clients:  cmap.New(),
	}
	return mgr
}

func (s *RemoteMgr) Start(actorSystem *actor.ActorSystem) error {
	s.actorSystem = actorSystem

	listener := network.StartTcpListen(s.actorSystem.Address(), func() network.ICodec { return &network.StreamCodec{} }, func() network.INetHandler { return &remoteHandler{remote: s} })
	err := listener.Start()
	if err != nil {
		return err
	}

	s.ping, err = actor_msg.NewNetActorMessage("", "", "", "$ping", nil).Marshal()
	if err != nil {
		return err
	}

	s.regist, err = actor_msg.NewNetActorMessage("", "", "", "$regist", []byte(s.actorSystem.Address())).Marshal()
	if err != nil {
		return err
	}

	s.listener = listener
	tools.GoEngine(s.keepAlive)

	s.actorSystem.RegistCmd("", "remoteinfo", s.remoteinfo)
	return err
}

func (s *RemoteMgr) Stop() {
	if s.stop.CAS(0, 1) {
		if s.listener != nil {
			s.listener.Stop()
		}
		s.clients.IterCb(func(key string, v interface{}) { v.(network.INetClient).Stop() })
	}
}

func (s *RemoteMgr) NewClient(host string) {
	c := network.NewTcpClient(host, func() network.ICodec { return &network.StreamCodec{} }, func() network.INetHandler { return &remoteHandler{remote: s, peerHost: host} })
	s.clients.Set(host, c)
	c.Start(true)
}

func (s *RemoteMgr) StopClient(host string) {
	if c, ok := s.clients.Pop(host); ok {
		c.(network.INetClient).Stop()
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
			s.sessions.IterCb(func(key string, v interface{}) { v.(*remoteHandler).SendMsg(s.ping) })
		}
	}
}

// 远端actor发送消息
func (s *RemoteMgr) Send(addr string, sourceId, targetId, requestId string, actMsg gproto.Message) error {
	session, ok := s.sessions.Get(addr)
	if !ok {
		return errors.New("remote addr not found")
	}

	//TODO 开协程处理？
	data, err := gproto.Marshal(actMsg)
	if err != nil {
		return err
	}

	msg := actor_msg.NewNetActorMessage(sourceId, targetId, requestId, tools.MsgName(actMsg), data)
	data, err = msg.Marshal()
	if err != nil {
		return err
	}
	return session.(*remoteHandler).SendMsg(data)
}

///////////////////////////////////////// remoteHandler /////////////////////////////////////////////
type remoteHandler struct {
	network.BaseNetHandler

	remote   *RemoteMgr
	peerHost string
	logger   *log.Logger
}

func (s *remoteHandler) OnSessionCreated() {
	s.logger = log.NewWithDefaultAndLogger(logger, map[string]interface{}{"local": s.LocalAddr(), "remote": s.RemoteAddr(), "session": s.Id()})
	s.SendMsg(s.remote.regist)
}

func (s *remoteHandler) OnSessionClosed() {
	if len(s.peerHost) > 0 {
		s.remote.sessions.Remove(s.peerHost)
		s.remote.actorSystem.DispatchEvent("$remoteHandler", &actor.Ev_delSession{Host: s.peerHost})
	}
}

func (s *remoteHandler) OnRecv(data []byte) {
	if s.remote == nil {
		s.Stop()
		return
	}

	msg := &actor_msg.ActorMessage{}
	err := msg.Unmarshal(data)
	if err != nil {
		s.logger.Error("unmarshal msg failed")
		return
	}

	if msg.MsgName == "$regist" {
		s.peerHost = string(msg.Data)
		s.remote.sessions.Set(s.peerHost, s)
		s.remote.actorSystem.DispatchEvent("$remoteHandler", &actor.Ev_newSession{Host: s.peerHost})
	} else if msg.MsgName == "$ping" {
		//do nothing
		//s.logger.Debug("recv ping")
	} else {
		if s.peerHost == "" {
			s.logger.KV("msg", msg.MsgName).Error("has not regist")
			return
		}

		tp, err := tools.FindMsgByName(msg.MsgName)
		if err != nil {
			s.logger.KV("MsgName", msg.MsgName).KV("err", err).Error("msg name not find")
			return
		}

		actMsg := tp.New().Interface().(gproto.Message)
		if err = gproto.Unmarshal(msg.Data, actMsg); err != nil {
			s.logger.KV("MsgName", msg.MsgName).KV("err", err).Error("Unmarshal failed")
			return
		}
		s.remote.actorSystem.Send(msg.SourceId, msg.TargetId, msg.RequestId, actMsg)
	}
}
