package remote_grpc

import (
	"fmt"
	//"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/atomic"

	"github.com/wwj31/dogactor/actor/cluster/remote_provider"
	"github.com/wwj31/dogactor/actor/cluster/remote_provider/remote_grpc/internal"
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

// 管理所有远端的session
type RemoteMgr struct {
	remoteHandler remote_provider.RemoteHandler
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

func (s *RemoteMgr) Start(actorSystem remote_provider.RemoteHandler) error {
	s.remoteHandler = actorSystem

	listener, err := internal.NewServer(s.remoteHandler.Address(), func() internal.IHandler { return &remoteHandler{remote: s} }, s)
	if err != nil {
		return err
	}

	s.regist = actor_msg.NewNetActorMessage("", "", "", "$regist", []byte(s.remoteHandler.Address()))
	s.listener = listener

	//s.remoteHandler.RegistCmd("", "remoteinfo", s.remoteinfo)
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
	c := internal.NewClient(host, func() internal.IHandler { return &remoteHandler{remote: s, peerHost: host} })
	s.clients.Set(host, c)
	_ = c.Start(true)
}

func (s *RemoteMgr) StopClient(host string) {
	if c, ok := s.clients.Pop(host); ok {
		c.(*internal.Client).Stop()
	}
}

// 远端actor发送消息
func (s *RemoteMgr) SendMsg(addr string, sourceId, targetId, requestId string, msg proto.Message) error {
	session, ok := s.sessions.Get(addr)
	if !ok {
		return fmt.Errorf("remote addr not found %v reqId:%v", addr, requestId)
	}

	//TODO 开协程处理？
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	netMsg := actor_msg.NewNetActorMessage(sourceId, targetId, requestId, tools.MsgName(msg), data)
	return session.(*remoteHandler).Send(netMsg)
}

///////////////////////////////////////// remoteHandler /////////////////////////////////////////////
type remoteHandler struct {
	internal.BaseHandler

	remote   *RemoteMgr
	peerHost string
	logger   *log.Logger
}

func (s *remoteHandler) OnSessionCreated() {
	s.logger = log.NewWithDefaultAndLogger(logger, map[string]interface{}{"session": s.Id()})
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
			s.logger.KV("msg", msg.MsgName).Error("has not regist")
			return
		}

		tp, err := tools.FindMsgByName(msg.MsgName)
		if err != nil {
			s.logger.KV("msgName", msg.MsgName).KV("err", err).Error("msg name not find")
			return
		}

		actMsg := tp.New().Interface().(proto.Message)
		if err = proto.Unmarshal(msg.Data, actMsg); err != nil {
			s.logger.KV("msgName", msg.MsgName).KV("err", err).Error("Unmarshal failed")
			return
		}
		s.remote.remoteHandler.OnSessionRecv(msg.SourceId, msg.TargetId, msg.RequestId, actMsg)
	}
}
