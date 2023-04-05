package actor

import (
	"time"

	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type mailBox struct {
	ch             chan Message
	lastMsgName    string
	processingTime time.Duration
}

// recording record slow processed of message
func (m *mailBox) recording(t time.Time, msgName string) {
	dur := tools.Now().Sub(t)
	m.processingTime = dur
	m.lastMsgName = msgName

	if dur > 1500*time.Millisecond {
		log.SysLog.Warnw("too long to process time", "msg", msgName, "duration", dur)
	}
}

func (m *mailBox) empty() bool {
	return len(m.ch) == 0
}
