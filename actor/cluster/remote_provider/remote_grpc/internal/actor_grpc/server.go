package actor_grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

func NewServer(host string, newHandler func() IHandler, ext interface{}) (*Server, error) {
	ln, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}

	kaep := grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{PermitWithoutStream: true})
	kasp := grpc.KeepaliveParams(keepalive.ServerParameters{Time: 15 * time.Second, Timeout: 5 * time.Second})
	gs := grpc.NewServer(kaep, kasp)

	server := &Server{Server: gs, newHandler: newHandler, ext: ext}
	actor_msg.RegisterActorServiceServer(gs, server)

	log.KV("host", host).Info("grpc server start")
	tools.GoEngine(func() {
		err := gs.Serve(ln)
		if err != nil {
			log.KV("host", host).KV("error", err).Error("grpc server failed")
		}
	})

	return server, nil
}

type Server struct {
	*grpc.Server
	newHandler func() IHandler
	ext        interface{}
}

func (this *Server) Receive(stream actor_msg.ActorService_ReceiveServer) error {
	session := NewSession(stream, this.newHandler())
	session.ext.Store("ext", this.ext)
	session.start(true)
	return nil
}
