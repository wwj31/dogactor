package remote_provider

import (
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
)

type RemoteHandler interface {
	Address() string
	OnSessionClosed(peerHost string)
	OnSessionOpened(peerHost string)
	OnSessionRecv(msg *actor_msg.ActorMessage)
}
