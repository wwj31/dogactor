package actorerr

import (
	"errors"
)

//errors

var (
	//error of event
	RegisterEventErr = errors.New("RegistEvent only accept ptr param")
	CancelEventErr   = errors.New("CancelEvent only accept ptr param")
	DispatchEventErr = errors.New("DispatchEvent only accept ptr param")

	//error of actor system
	ActorSystemOptionErr   = errors.New("actor system option run failed")
	RegistClusterErr       = errors.New("regist cluster error")
	RegisterActorSystemErr = errors.New("actor system has stopped")
	RegisterActorSameIdErr = errors.New("actor with same id")
	ProtoMarshalErr 	   = errors.New("actor system send proto marshal failed")
	ActorNotFoundErr       = errors.New("local actor not found")
	ActorPushMsgErr        = errors.New("push a nil msg to actor")
	ActorUnimplemented     = errors.New("actor has no implemented OnHandleRequest")
)
