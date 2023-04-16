package remote

type Handler interface {
	Address() string
	OnSessionClosed(peerHost string)
	OnSessionOpened(peerHost string)
	OnSessionRecv(msg *innermsg.ActorMessage)
}
