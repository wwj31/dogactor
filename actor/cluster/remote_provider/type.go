package remote_provider

import "github.com/golang/protobuf/proto"

type RemoteHandler interface {
	Address() string
	OnSessionClosed(peerHost string)
	OnSessionOpened(peerHost string)
	OnSessionRecv(sourceId, targetId, requestId string, msg proto.Message)
}
