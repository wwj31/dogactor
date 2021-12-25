package actor

import (
	"errors"
	"fmt"
	"github.com/wwj31/dogactor/log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/tools"
)

var (
	_id         = time.Now().UnixNano()
	requestPool = sync.Pool{New: func() interface{} { return &request{} }}
)

const DefaultTimeout = time.Second * 5

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
		log.SysLog.Errorw("repeated set handle request id", "requestId", s.id)
		return
	}
	s.fn = fn
	if s.result != nil || s.err != nil {
		s.fn(s.result, s.err)
		requestPool.Put(s)
	}
}

func (s *actor) Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request) {
	req = requestPool.Get().(*request)
	req.id = requestId(s.id, targetId, s.system.Address())
	req.sourceId = s.ID()
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

	req.timeoutId = s.AddTimer(tools.UUID(), tools.NowTime()+int64(interval), func(dt int64) {
		expect.Nil(s.Response(req.id, &actor_msg.RequestDeadLetter{Err: "Request timeout"}))
	})
	return req
}

// RequestWait sync request
func (s *actor) RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (resp interface{}, err error) {
	var t time.Duration
	if len(timeout) > 0 && timeout[0] > 0 {
		t = timeout[0]
	}

	waitRsp := make(chan result)
	waiter := New(
		"wait_"+tools.UUID(),
		&waitActor{c: waitRsp, msg: msg, targetId: targetId, timeout: t},
		SetLocalized(),
	)
	expect.Nil(s.System().Regist(waiter))

	// wait to result
	r := <-waitRsp
	return r.result, r.err
}

// Response response a result
func (s *actor) Response(requestId string, msg interface{}) error {
	reqSourceId, _, _, ok := ParseRequestId(requestId)
	if !ok {
		return fmt.Errorf("error requestId:%v", requestId)
	}
	return s.system.Send(s.id, reqSourceId, requestId, msg)
}

// process to Response msg
func (s *actor) doneRequest(requestId string, resp interface{}) {
	req, ok := s.requests[requestId]
	if !ok {
		log.SysLog.Warnw("can not find request", "requestId", requestId, "actorId", s.id)
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
		requestPool.Put(req)
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
