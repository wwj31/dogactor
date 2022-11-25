package network

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type OptionClient func(l *TcpClient)

type TcpClient struct {
	addr    string
	running int32

	newCodec    func() DecodeEncoder
	handlersFun []func() SessionHandler

	session    *TcpSession
	reconTimes int
	recon      chan struct{}

	connMux sync.Mutex
}

func NewTcpClient(addr string, newCodec func() DecodeEncoder, opt ...OptionClient) Client {
	c := &TcpClient{
		addr:        addr,
		running:     1,
		newCodec:    newCodec,
		handlersFun: make([]func() SessionHandler, 0),
		recon:       make(chan struct{}, 1),
	}
	h := func() SessionHandler {
		return &tcpReconnectHandler{reconnect: c.recon}
	}
	c.AddHandler(h)

	for _, f := range opt {
		f(c)
	}
	return c
}

// AddHandler add handler to tail of list
func (s *TcpClient) AddHandler(hander func() SessionHandler) {
	s.handlersFun = append(s.handlersFun, hander)
}

func (s *TcpClient) Start(reconnect bool) error {
	if reconnect {
		go s.reconnect()
	}
	return s.connect()
}

func (s *TcpClient) SendMsg(data []byte) error {
	if s.session != nil {
		return s.session.SendMsg(data)
	}
	return errors.New("has not connect")
}

func (s *TcpClient) Stop() {
	if atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		s.connMux.Lock()
		defer s.connMux.Unlock()

		log.SysLog.Infow("tcp client stop", "addr", s.addr)
		s.session.Stop()
	}
}

func (s *TcpClient) isRunning() bool { return atomic.LoadInt32(&s.running) == 1 }
func (s *TcpClient) connect() error {
	s.connMux.Lock()
	defer s.connMux.Unlock()

	if !s.isRunning() {
		return errors.New("tcp client has stopped")
	}

	conn, err := net.DialTimeout("tcp", s.addr, time.Second)
	if err != nil {
		return err
	}

	var handlers []SessionHandler
	for _, f := range s.handlersFun {
		handlers = append(handlers, f())
	}
	s.session = newTcpSession(conn, s.newCodec(), handlers...)
	s.session.start()
	return nil
}

func (s *TcpClient) reconnect() {
	for range s.recon {
		if !s.isRunning() {
			return
		}

		if err := s.connect(); err != nil {
			log.SysLog.Warnw("tcp client try reconnect",
				"recon times", s.reconTimes, "addr", s.addr, "connect err", err)

			time.Sleep(time.Second * time.Duration(s.reconTimes))
			s.reconTimes = tools.Min(20, s.reconTimes+1)
			if len(s.recon) == 0 {
				s.recon <- struct{}{}
			}
			continue
		}

		// success reconnect
		s.reconTimes = 0
	}
}
