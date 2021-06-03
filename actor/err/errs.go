package err

import "fmt"

//errors

var (
	//error of event
	RegisterEventErr = fmt.Errorf("RegistEvent only accept ptr param")
	CancelEventErr   = fmt.Errorf("CancelEvent only accept ptr param")
	DispatchEventErr = fmt.Errorf("DispatchEvent only accept ptr param")
	EventHasNotErr   = fmt.Errorf("system event has not been set")

	//error of actor system
	ActorSystemOptionErr   = fmt.Errorf("actor system option run failed")
	RegistClusterErr       = fmt.Errorf("regist cluster error")
	ClusterNeedEventErr    = fmt.Errorf("cluster need event")
	RegisterActorSystemErr = fmt.Errorf("actor system has stopped")
	RegisterActorSameIdErr = fmt.Errorf("actor with same id")
	ActorNotFoundErr       = fmt.Errorf("local actor not found")
	ActorPushMsgErr        = fmt.Errorf("push a nil msg to actor")
	ActorUnimplemented     = fmt.Errorf("actor has no implemented HandleRequest")
)
