package remote_tcp

import (
	"errors"
	"github.com/golang/protobuf/proto"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/wwj31/dogactor/actor/cluster/remote_provider"
	"github.com/wwj31/dogactor/expect"
	"go.uber.org/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/log"
	"github.com/wwj31/dogactor/network"
	"github.com/wwj31/dogactor/tools"
)

// 管理所有远端的session
type RemoteMgr struct {
	remoteHandler remote_provider.RemoteHandler
	listener      network.INetListener

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

func (s *RemoteMgr) Start(h remote_provider.RemoteHandler) error {
	s.remoteHandler = h

	listener := network.StartTcpListen(s.remoteHandler.Address(),
		func() network.ICodec { return &network.StreamCodec{} },
		func() network.INetHandler { return &remoteHandler{remote: s}},
	)

	err := listener.Start()
	if err != nil {
		return err
	}

	s.ping, err = actor_msg.NewNetActorMessage("", "", "", "$ping", nil).Marshal()
	if err != nil {
		return err
	}

	s.regist, err = actor_msg.NewNetActorMessage("", "", "", "$regist", []byte(s.remoteHandler.Address())).Marshal()
	if err != nil {
		return err
	}

	s.listener = listener
	go s.keepAlive()

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
	c := network.NewTcpClient(host, func() network.ICodec { return &network.StreamCodec{} })
	c.AddLast(func() network.INetHandler { return &remoteHandler{remote: s, peerHost: host} })
	s.clients.Set(host, c)
	expect.Nil(c.Start(true))
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
func (s *RemoteMgr) SendMsg(addr string, sourceId, targetId, requestId string, msg proto.Message) error {
	session, ok := s.sessions.Get(addr)
	if !ok {
		return errors.New("remote addr not found")
	}

	//TODO 开协程处理？
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	netMsg := actor_msg.NewNetActorMessage(sourceId, targetId, requestId, tools.MsgName(msg), data)
	defer netMsg.Free()
	data, err = netMsg.Marshal()
	if err != nil {
		return err
	}
	return session.(*remoteHandler).SendMsg(data)
}

///////////////////////////////////////// remoteHandler /////////////////////////////////////////////
type remoteHandler struct {
	network.INetSession
	remote   *RemoteMgr
	peerHost string
}

func (s *remoteHandler) OnSessionCreated(sess network.INetSession) {
	s.INetSession = sess
	err := sess.SendMsg(s.remote.regist)
	if err != nil {
		log.SysLog.Errorw("OnSessionCreated error","err",err)
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

	msg := &actor_msg.ActorMessage{}
	err := msg.Unmarshal(data)
	if err != nil {
		log.SysLog.Errorf("unmarshal msg failed","err",err)
		return
	}

	if msg.MsgName == "$regist" {
		s.peerHost = string(msg.Data)
		s.remote.sessions.Set(s.peerHost, s)
		s.remote.remoteHandler.OnSessionOpened(s.peerHost)
	} else if msg.MsgName == "$ping" {
		//do nothing
		//s.logger.Debug("recv ping")
	} else {
		if s.peerHost == "" {
			log.SysLog.Errorf("has not regist","msg", msg.MsgName)
			return
		}

		tp, err := tools.FindMsgByName(msg.MsgName)
		if err != nil {
			log.SysLog.Errorf("msg name not find","err", err,"MsgName", msg.MsgName)
			return
		}

		actMsg := tp.New().Interface().(proto.Message)
		if err = proto.Unmarshal(msg.Data, actMsg); err != nil {
			log.SysLog.Errorf("Unmarshal failed","err", err,"MsgName", msg.MsgName)
			return
		}
		s.remote.remoteHandler.OnSessionRecv(msg.SourceId, msg.TargetId, msg.RequestId, actMsg)
	}
}
