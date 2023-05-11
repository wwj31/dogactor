package conntcp

import (
	"errors"
	"fmt"
	"net"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/atomic"

	"github.com/wwj31/dogactor/actor/cluster/fullmesh/remote"
	"github.com/wwj31/dogactor/actor/internal/innermsg"
	"github.com/wwj31/dogactor/expect"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/network"
)

type RemoteMgr struct {
	addr          string
	exceptPort    []int
	remoteHandler remote.Handler
	listener      network.Listener

	stop     atomic.Int32
	sessions cmap.ConcurrentMap //host=>session
	clients  cmap.ConcurrentMap //host=>client

	ping     []byte
	registry []byte
}

func NewRemoteMgr(exceptPort ...int) *RemoteMgr {
	mgr := &RemoteMgr{
		exceptPort: exceptPort,
		sessions:   cmap.New(),
		clients:    cmap.New(),
	}
	return mgr
}

func (s *RemoteMgr) Addr() string {
	return s.addr
}

func (s *RemoteMgr) Start(h remote.Handler) error {
	s.remoteHandler = h

	var (
		port int
		err  error
	)

	s.listener = network.StartTcpListen("0.0.0.0:0",
		func() network.DecodeEncoder { return &network.StreamCode{} },
		func() network.SessionHandler { return &remoteHandler{remote: s} },
	)

	err = s.listener.Start(s.exceptPort...)
	if err != nil {
		return err
	}

	port = s.listener.Port()

	ping := innermsg.NewActorMessage() // ping message
	ping.MsgName = "$ping"
	s.ping, err = ping.Marshal()
	if err != nil {
		return err
	}

	reg := innermsg.NewActorMessage() // registry message
	reg.MsgName = "$registry"

	ip, ipErr := getIntranetIP()
	if ipErr != nil {
		return ipErr
	}

	s.addr = fmt.Sprintf("%v:%v", ip, port)
	reg.Data = []byte(s.addr)
	s.registry, err = reg.Marshal()
	if err != nil {
		return err
	}

	go s.keepAlive()

	log.SysLog.Infow("fullMesh listen success", "addr", s.addr)
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
	c.AddHandler(func() network.SessionHandler { return &remoteHandler{remote: s, peerHost: host} })
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
				handler, _ := v.(*remoteHandler)
				err := v.(*remoteHandler).SendMsg(s.ping)
				if err != nil {
					log.SysLog.Errorw("keepAlive SendMsg",
						"err", err,
						"handler", handler.RemoteIP())
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

// /////////////////////////////////////// remoteHandler /////////////////////////////////////////////
type remoteHandler struct {
	network.Session
	remote   *RemoteMgr
	peerHost string
}

func (s *remoteHandler) OnSessionCreated(sess network.Session) {
	s.Session = sess
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

	msg := innermsg.NewActorMessage() // remote recv message
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

// getIntranetIP 获取本机内网IP，只返回其中一个
func getIntranetIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("cannot find intra net")
}
