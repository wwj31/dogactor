package internal

type EvClusterUpdate struct {
	ActorId string
	Host    string
	Add     bool
}

type EvSessionClosed struct {
	PeerHost string
}

type EvSessionOpened struct {
	PeerHost string
}

type ReqMsgDrain struct{}
type RespMsgDrain struct{ Err error }
