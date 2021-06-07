package actor

import (
	"errors"
	"fmt"
	"github.com/wwj31/godactor/expect"
	"github.com/wwj31/godactor/tools"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wwj31/godactor/actor/internal/actor_msg"
)

var _id = time.Now().UnixNano()

const DefaultTimeout = time.Second * 30

type request struct {
	id       string
	sourceId string
	targetId string

	result    interface{}
	err       error
	fn        func(resp interface{}, err error)
	timeoutId int64
}

func (s *request) Handle(fn func(resp interface{}, err error)) {
	if s.fn != nil {
		logger.KV("requestId", s.id).ErrorStack(3, "repeated set handle request id")
		return
	}
	s.fn = fn
	if s.result != nil || s.err != nil {
		s.fn(s.result, s.err)
	}
}

func (s *actor) Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request) {
	req = s.newRequest(targetId)
	e := s.system.Send(s.id, targetId, req.id, msg)
	if req.err != nil {
		req.err = e
		return req
	}

	interval := DefaultTimeout
	if len(timeout) > 0 {
		interval = timeout[0]
	}

	if interval > 0 {
		req.timeoutId = s.AddTimer(interval, 1, func(dt int64) {
			expect.Nil(s.Response(req.id, &actor_msg.RequestDeadLetter{Err: "Request timeout"}))
		})
	}
	return req
}

//谨慎使用，可能带来死锁问题
type waitActor struct {
	ActorHanlerBase
}

func (s *waitActor) Stop() bool {
	return false
}

func (s *actor) RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (result interface{}, err error) {
	//新起actor等待结果
	waiter := New(tools.UUID(), &waitActor{}, SetLocalized())
	expect.Nil(s.System().Regist(waiter))

	//超时设定
	interval := DefaultTimeout
	if len(timeout) > 0 {
		interval = timeout[0]
	}

	req := waiter.Request(targetId, msg, -1)

	waiter.AddTimer(interval, 1, func(dt int64) {
		expect.Nil(waiter.Response(req.id, &actor_msg.RequestDeadLetter{Err: "RequestWait timeout"}))
	})

	//发出请求，并阻塞等待结果
	respC := make(chan struct{})
	req.Handle(func(resp interface{}, e error) {
		result = resp
		err = e
		respC <- struct{}{}
		waiter.Stop()
	})

	<-respC // 阻塞等待waiter返回结果
	return
}

func (s *actor) Response(requestId string, msg interface{}) error {
	reqSourceId, _, _, ok := ParseRequestId(requestId)
	if !ok {
		return fmt.Errorf("error requestId:%v", requestId)
	}
	return s.system.Send(s.id, reqSourceId, requestId, msg)
}

func (s *actor) newRequest(targetId string) (req *request) {
	req = &request{
		id:       requestId(s.id, targetId, s.system.Address()),
		sourceId: s.GetID(),
		targetId: targetId,
		result:   nil,
		err:      nil,
		fn:       nil,
	}
	s.requests[req.id] = req
	return req
}

func (s *actor) doneRequest(requestId string, resp interface{}) {
	req, ok := s.requests[requestId]
	if !ok {
		s.logger.KV("requestId", requestId).Warn("can not find request")
		return
	}

	s.CancelTimer(req.timeoutId)
	delete(s.requests, requestId)

	switch r := resp.(type) {
	case *actor_msg.RequestDeadLetter:
		req.err = errors.New(r.Err)
	case error:
		req.err = r
	default:
		req.result = resp
	}

	if req.fn != nil {
		req.fn(req.result, req.err)
	}
}

func requestId(actorId, targetId, sourceAddr string) string {
	return fmt.Sprintf("%s$%d$%s@%s", actorId, atomic.AddInt64(&_id, 1), targetId, sourceAddr)
}

func ParseRequestId(requestId string) (source string, target string, addr string, ok bool) {
	if len(requestId) == 0 {
		return
	}
	strs := strings.Split(requestId, "@")
	if len(strs) != 2 {
		return
	}
	addr = strs[1]
	strs = strings.Split(strs[0], "$")
	if len(strs) != 3 {
		return
	}
	ok = true
	return strs[0], strs[2], addr, ok
}
