package actor

import (
	"github.com/wwj31/dogactor/actor/internal/actor_msg"
	"github.com/wwj31/dogactor/log"
	"time"
)

type mailBox struct {
	ch             chan actor_msg.Message
	lastMsgName    string
	processingTime time.Duration
}

// recording record slow processed of message
func (m *mailBox) recording(t time.Time, msgName string) {
	dur := time.Now().Sub(t)
	m.processingTime = dur
	m.lastMsgName = msgName

	if dur > 150*time.Millisecond {
		log.SysLog.Warnw("too long to process time", "msg", msgName, "duration", dur)
	}
}
