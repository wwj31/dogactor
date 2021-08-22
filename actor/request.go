package actor

import (
	"errors"
	"fmt"
	"github.com/wwj31/godactor/expect"
	"github.com/wwj31/godactor/log"
	"github.com/wwj31/godactor/tools"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwj31/godactor/actor/internal/actor_msg"
)

var (
	_id          = time.Now().UnixNano()
	request_pool = sync.Pool{New: func() interface{} { return &request{} }}
)

const DefaultTimeout = time.Second * 30

type request struct {
	id       string
	sourceId string
	targetId string

	result    interface{}
	err       error
	fn        func(resp interface{}, err error)
	timeoutId string
}

func (s *request) Handle(fn func(resp interface{}, err error)) {
	if s.fn != nil {
		logger.KV("requestId", s.id).ErrorStack(3, "repeated set handle request id")
		return
	}
	s.fn = fn
	if s.result != nil || s.err != nil {
		s.fn(s.result, s.err)
		request_pool.Put(s)
	}
}

func (s *actor) Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request) {
	req = request_pool.Get().(*request)
	req.id = requestId(s.id, targetId, s.system.Address())
	req.sourceId = s.GetID()
	req.targetId = targetId
	req.result = nil
	req.err = nil
	req.fn = nil
	s.requests[req.id] = req

	err := s.system.Send(s.id, targetId, req.id, msg)
	if err != nil {
		req.err = err
		return req
	}

	interval := DefaultTimeout
	if len(timeout) > 0 && timeout[0] > 0 {
		interval = timeout[0]
	}

	req.timeoutId = s.AddTimer(tools.UUID(), interval, func(dt int64) {
		expect.Nil(s.Response(req.id, &actor_msg.RequestDeadLetter{Err: "Request timeout"}))
	})
	return req
}

// 同步请求结果
func (s *actor) RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (resp interface{}, err error) {
	if atomic.LoadInt32(&s.asyncStop) == 1 {
		return
	}
	waitRsp := make(chan result)
	waiter := New(tools.UUID(), &waitActor{c: waitRsp, msg: msg, targetId: targetId}, SetLocalized())
	expect.Nil(s.System().Regist(waiter))

	// 阻塞等待waiter返回结果
	r := <-waitRsp
	return r.result, r.err
}

func (s *actor) Response(requestId string, msg interface{}) error {
	reqSourceId, _, _, ok := ParseRequestId(requestId)
	if !ok {
		return fmt.Errorf("error requestId:%v", requestId)
	}
	return s.system.Send(s.id, reqSourceId, requestId, msg)
}

// 收到response消息的处理逻辑
func (s *actor) doneRequest(requestId string, resp interface{}) {
	req, ok := s.requests[requestId]
	if !ok {
		s.logger.KVs(log.Fields{"requestId": requestId, "actor": s.id}).Warn("can not find request")
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
		request_pool.Put(req)
	}
}

// actorId-incId-targetId#sourceAddr
func requestId(actorId, targetId, sourceAddr string) string {
	return fmt.Sprintf("%s@%d@%s#%s", actorId, atomic.AddInt64(&_id, 1), targetId, sourceAddr)
}

func ParseRequestId(requestId string) (sourceId string, targetId string, sourceAddr string, ok bool) {
	if len(requestId) == 0 {
		return
	}
	strs := strings.Split(requestId, "#")
	if len(strs) != 2 {
		return
	}
	sourceAddr = strs[1]
	strs = strings.Split(strs[0], "@")
	if len(strs) != 3 {
		return
	}
	ok = true
	return strs[0], strs[2], sourceAddr, ok
}
