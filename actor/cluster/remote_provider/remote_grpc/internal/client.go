package internal

import (
	"context"
	"errors"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"sync"
	"time"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

var (
	NotRunning = errors.New("client has stopped")
)

func NewClient(host string, newHandler func() IHandler) *Client {
	c := &Client{
		host:             host,
		sessionReconnect: make(chan struct{}, 1),
	}
	c.newHandler = func() IHandler { return &ClientHander{client: c, handler: newHandler()} }
	return c
}

type Client struct {
	host             string
	conn             *grpc.ClientConn
	newHandler       func() IHandler
	reconnectTimes   int
	running          atomic.Int32
	sessionReconnect chan struct{}
	connMux          sync.Mutex
}

func (g *Client) Start(reconnect bool) error {
	if reconnect {
		tools.GoEngine(g.reconnect)
	} else {
		return g.connect()
	}
	return nil
}

func (g *Client) Stop() {
	if g.running.CAS(0, 1) {
		g.connMux.Lock()
		defer g.connMux.Unlock()
		if g.conn != nil {
			g.conn.Close()
		}
	}
}

func (g *Client) IsRunning() bool { return g.running.Load() == 0 }

func (g *Client) connect() error {
	g.connMux.Lock()
	defer g.connMux.Unlock()

	if !g.IsRunning() {
		return NotRunning
	}

	//kacp := grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 15 * time.Second, Timeout: 5 * time.Second, PermitWithoutStream: true})
	conn, err := grpc.Dial(g.host, grpc.WithInsecure())
	if conn == nil {
		log.KV("host", g.host).KV("error", err).Error("connect grpc remote error")
		return err
	}

	stream, err := actor_msg.NewActorServiceClient(conn).Receive(context.Background())
	if stream == nil {
		log.KV("host", g.host).KV("error", err).Error("connect grpc remote error")
		return err
	}
	g.conn = conn
	cli := NewSession(stream, g.newHandler())
	cli.start(false)
	return nil
}

func (g *Client) reconnect() {
	g.sessionReconnect <- struct{}{}
	for {
		select {
		case <-g.sessionReconnect:
			if !g.IsRunning() {
				return
			}

			if err := g.connect(); err == nil {
				g.reconnectTimes = 0
				break
			}

			time.Sleep(time.Second * time.Duration(g.reconnectTimes))
			g.reconnectTimes++
			g.sessionReconnect <- struct{}{}
		}
	}
}

type ClientHander struct {
	BaseHandler
	client  *Client
	handler IHandler
}

func (s *ClientHander) OnSessionCreated() {
	s.handler.setSession(s.BaseHandler.Session)
	s.handler.OnSessionCreated()
}

func (s *ClientHander) OnSessionClosed() {
	s.handler.OnSessionClosed()
	s.client.sessionReconnect <- struct{}{}
}

func (s *ClientHander) OnRecv(msg *actor_msg.ActorMessage) {
	s.handler.OnRecv(msg)
}
