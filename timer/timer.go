package timer

import (
	"fmt"
	"time"
)

type Id = string

type (
	timer struct {
		id        Id
		startAt   time.Time
		endAt     time.Time
		callback  CallbackFn
		execCount int
		remove    bool // soft remove flags
		index     int
	}
)

func (t *timer) spareCount() bool {
	return t.execCount > 0 || t.execCount < 0
}

func (t *timer) consumeCount(count int) {
	if t.execCount > 0 {
		t.execCount -= min(t.execCount, count)
	}
}

func (t *timer) String() string {
	return fmt.Sprintf("[endAt:%v index:%v] ", t.endAt, t.index)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
