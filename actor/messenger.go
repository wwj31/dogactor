package actor

import (
	"errors"
	"fmt"
	"time"

	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/internal/innermsg"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type communication struct {
	*Base

	// record Request msg and delete after done
	requests map[RequestId]*request
}

func newCommunication(base *Base) *communication {
	cm := &communication{
		Base:     base,
		requests: make(map[RequestId]*request),
	}

	atr := base.Actor.(*actor)
	atr.appendHandler(func(message Message) bool {
		reqId := RequestId(message.GetRequestId())
		reqSourceId, _, _, _ := reqId.Parse()
		if cm.ID() == reqSourceId {
			//handle Response
			cm.doneRequest(message.GetRequestId(), message.Payload())
			return false
		}
		return true
	})
	return cm
}

func (s *communication) Send(targetId string, msg any) error {
	return s.System().Send(s.ID(), targetId, "", msg)
}

func (s *communication) Request(targetId string, msg any, timeout ...time.Duration) (req *request) {
	req = requestPool.Get().(*request)
	req.id = requestId(s.ID(), targetId, s.System().Addr)
	req.sourceId = s.ID()
	req.targetId = targetId
	req.result.data = nil
	req.result.err = nil
	req.fn = nil
	s.requests[req.id] = req

	err := s.System().Send(s.ID(), targetId, req.id, msg)
	if err != nil {
		req.result.err = err
		return req
	}

	interval := DefaultTimeout
	if len(timeout) > 0 && timeout[0] > 0 {
		interval = timeout[0]
	}

	req.timeoutId = s.AddTimer(tools.XUID(), tools.Now().Add(interval), func(dt time.Duration) {
		expect.Nil(s.Response(req.id.String(), &innermsg.RequestDeadLetter{Err: "Request timeout"}))
	})
	return req
}

// RequestWait sync request
func (s *communication) RequestWait(targetId string, msg any, timeout ...time.Duration) (resp any, err error) {
	if targetId == s.ID() {
		return nil, actorerr.ActorForbiddenToCallOneself
	}

	var t time.Duration
	if len(timeout) > 0 && timeout[0] > 0 {
		t = timeout[0]
	}

	waitRsp := make(chan result)
	expect.Nil(s.Send(s.System().requestWaiter, &requestWait{
		targetId: targetId,
		timeout:  t,
		msg:      msg,
		response: waitRsp,
	}))

	// wait to result
	res := <-waitRsp
	return res.data, res.err
}

// Response response a result
func (s *communication) Response(requestId string, msg interface{}) error {
	reqId := RequestId(requestId)
	reqSourceId, _, _, ok := reqId.Parse()
	if !ok {
		return fmt.Errorf("error requestId:%v", requestId)
	}
	return s.System().Send(s.ID(), reqSourceId, reqId, msg)
}

// process to Response msg
func (s *communication) doneRequest(requestId string, resp interface{}) {
	req, ok := s.requests[RequestId(requestId)]
	if !ok {
		log.SysLog.Warnw("can not find request", "requestId", requestId, "actorId", s.ID())
		return
	}

	s.CancelTimer(req.timeoutId)
	delete(s.requests, RequestId(requestId))

	switch r := resp.(type) {
	case *innermsg.RequestDeadLetter:
		req.result.err = errors.New(r.Err)
	case error:
		req.result.err = r
	default:
		req.result.data = resp
	}

	if req.fn != nil {
		req.fn(req.result.data, req.result.err)
		requestPool.Put(req)
	}
}
