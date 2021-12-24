package internal

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/actor/log"
)

func NewServer(host string, newHandler func() Handler, ext interface{}) (*Server, error) {
	ln, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}

	kaep := grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{PermitWithoutStream: true})
	kasp := grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second, Timeout: 5 * time.Second})
	gs := grpc.NewServer(kaep, kasp)

	server := &Server{Server: gs, newHandler: newHandler, ext: ext}
	actor_msg.RegisterActorServiceServer(gs, server)

	log.SysLog.Infow("grpc server start","host", host)
	go func() {
		err := gs.Serve(ln)
		if err != nil {
			log.SysLog.Infow("grpc server failed","host", host,"error", err)
		}
	}()

	return server, nil
}

type Server struct {
	*grpc.Server
	newHandler func() Handler
	ext        interface{}
}

func (this *Server) Receive(stream actor_msg.ActorService_ReceiveServer) error {
	session := NewSession(stream, this.newHandler())
	session.ext.Store("ext", this.ext)
	session.start(true)
	return nil
}
