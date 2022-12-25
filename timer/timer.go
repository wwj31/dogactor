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
		callback CallbackFn
		times    int
		remove   bool // soft remove flags
		index     int
	}
)

func (t *timer) spareCount() bool {
	return t.times > 0 || t.times < 0
}

func (t *timer) consumeCount(count int) {
	if t.times > 0 {
		t.times -= min(t.times, count)
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
