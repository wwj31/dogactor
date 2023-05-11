package remote

import (
	"github.com/wwj31/dogactor/actor/internal/innermsg"
)

type Handler interface {
	OnSessionClosed(peerHost string)
	OnSessionOpened(peerHost string)
	OnSessionRecv(msg *innermsg.ActorMessage)
}
