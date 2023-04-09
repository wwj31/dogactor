package actor

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

var (
	_id         = tools.Now().UnixNano()
	requestPool = sync.Pool{New: func() interface{} { return &request{} }}
)

const DefaultTimeout = time.Second * 10

type (
	result struct {
		data interface{}
		err  error
	}

	request struct {
		id       RequestId
		sourceId Id
		targetId Id

		result    result
		fn        func(resp interface{}, err error)
		timeoutId string
	}
)

func (s *request) Handle(fn func(resp any, err error)) {
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

type RequestId string

// actorId@incId@targetId#sourceAddr
func requestId(actorId, targetId, sourceAddr string) RequestId {
	var builder strings.Builder
	builder.Grow(100)
	builder.WriteString(actorId)
	builder.WriteString("@")
	builder.WriteString(strconv.Itoa(int(atomic.AddInt64(&_id, 1))))
	builder.WriteString("@")
	builder.WriteString(targetId)
	builder.WriteString("#")
	builder.WriteString(sourceAddr)
	return RequestId(builder.String())
}

func (r RequestId) String() string {
	return string(r)
}

func (r RequestId) Valid() bool {
	return string(r) != ""
}

func (r RequestId) Parse() (sourceId string, targetId string, sourceAddr string, ok bool) {
	if len(r) == 0 {
		return
	}

	split := strings.Split(string(r), "#")
	if len(split) != 2 {
		return
	}

	sourceAddr = split[1]
	split = strings.Split(split[0], "@")
	if len(split) != 3 {
		return
	}

	return split[0], split[2], sourceAddr, true
}
