package network

import (
	"errors"
	"github.com/wwj31/dogactor/log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type OptionClient func(l *TcpClient)

type TcpClient struct {
	addr    string
	running int32

	newCodec    func() DecodeEncoder
	handlersFun []func() NetSessionHandler

	session     *TcpSession
	reconnTimes int
	reconn      chan struct{}

	connMux sync.Mutex
}

func NewTcpClient(addr string, newCodec func() DecodeEncoder, op ...OptionClient) Client {
	c := &TcpClient{
		addr:        addr,
		running:     1,
		newCodec:    newCodec,
		handlersFun: make([]func() NetSessionHandler, 0),
		reconn:      make(chan struct{}, 1),
	}
	h := func() NetSessionHandler {
		return &tcpReconnectHandler{reconnect: c.reconn}
	}
	c.AddLast(h)

	for _, f := range op {
		f(c)
	}
	return c
}

// add handler to last of list
func (s *TcpClient) AddLast(hander func() NetSessionHandler) {
	s.handlersFun = append(s.handlersFun, hander)
}

func (s *TcpClient) Start(reconnect bool) error {
	if reconnect {
		go s.reconnect()
	} else {
		return s.connect()
	}
	return nil
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

	handers := []NetSessionHandler{}
	for _, f := range s.handlersFun {
		handers = append(handers, f())
	}
	s.session = newTcpSession(conn, s.newCodec(), handers...)
	s.session.start()
	return nil
}

func (s *TcpClient) reconnect() {
	s.reconn <- struct{}{}
	for {
		select {
		case <-s.reconn:
			if !s.isRunning() {
				return
			}

			if err := s.connect(); err == nil {
				s.reconnTimes = 0
				break
			}

			time.Sleep(time.Second * time.Duration(s.reconnTimes))
			s.reconnTimes++
			s.reconn <- struct{}{}
		}
	}
}
