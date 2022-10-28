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

const DefaultTimeout = time.Second * 10

type (
	result struct {
		data interface{}
		err  error
	}

	request struct {
		id       string
		sourceId string
		targetId string

		result    result
		fn        func(resp interface{}, err error)
		timeoutId string
	}
)

func (s *request) Handle(fn func(resp interface{}, err error)) {
	if s.fn != nil {
		log.SysLog.Errorw("repeated set handle request id", "requestId", s.id)
		return
	}
	s.fn = fn
	if s.result.data != nil || s.result.err != nil {
		s.fn(s.result.data, s.result.err)
		requestPool.Put(s)
	}
}

func (s *actor) Request(targetId string, msg interface{}, timeout ...time.Duration) (req *request) {
	req = requestPool.Get().(*request)
	req.id = requestId(s.id, targetId, s.system.Address())
	req.sourceId = s.ID()
	req.targetId = targetId
	req.result.data = nil
	req.result.err = nil
	req.fn = nil
	s.requests[req.id] = req

	err := s.system.Send(s.id, targetId, req.id, msg)
	if err != nil {
		req.result.err = err
		return req
	}

	interval := DefaultTimeout
	if len(timeout) > 0 && timeout[0] > 0 {
		interval = timeout[0]
	}

	req.timeoutId = s.AddTimer(tools.XUID(), tools.Now().Add(interval), func(dt time.Duration) {
		expect.Nil(s.Response(req.id, &actor_msg.RequestDeadLetter{Err: "Request timeout"}))
	})
	return req
}

// RequestWait sync request
func (s *actor) RequestWait(targetId string, msg interface{}, timeout ...time.Duration) (resp interface{}, err error) {
	return s.System().RequestWait(targetId, msg, timeout...)
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

// actorId-incId-targetId#sourceAddr
func requestId(actorId, targetId, sourceAddr string) string {
	return fmt.Sprintf("%s@%d@%s#%s", actorId, atomic.AddInt64(&_id, 1), targetId, sourceAddr)
}

func ParseRequestId(requestId string) (sourceId string, targetId string, sourceAddr string, ok bool) {
	if len(requestId) == 0 {
		return
	}
	split := strings.Split(requestId, "#")
	if len(split) != 2 {
		return
	}
	sourceAddr = split[1]
	split = strings.Split(split[0], "@")
	if len(split) != 3 {
		return
	}
	ok = true
	return split[0], split[2], sourceAddr, ok
}
