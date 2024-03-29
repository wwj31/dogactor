package actorerr

import (
	"errors"
)

//errors

var (
	//errors about event

	CancelEventErr = errors.New("CancelEvent only accept ptr param")

	// errors about actor system

	ActorSystemOptionErr        = errors.New("actor system option run failed")
	RegisterClusterErr          = errors.New("register cluster error")
	RegisterActorSystemErr      = errors.New("actor system has stopped")
	RegisterActorSameIdErr      = errors.New("actor with same id")
	ProtoMarshalErr             = errors.New("actor system send proto marshal failed")
	ActorNotFoundErr            = errors.New("local actor not found")
	ActorMsgTypeCanNotRemoteErr = errors.New("msg type can not remote")
	ActorPushMsgErr             = errors.New("push a nil msg to actor")
	ActorForbiddenToCallOneself = errors.New("request wait not call one self")
)
