package timer

import (
	"fmt"
	"time"
)

type Id = string

type (
	timer struct {
		id          Id
		startAt     time.Time
		endAt       time.Time
		handler     callback
		repeatCount int
		remove      bool // soft remove flags
		index       int
	}
)

func (t *timer) spareCount() bool {
	return t.repeatCount > 0 || t.repeatCount < 0
}

func (t *timer) consumeCount(count int) {
	if t.repeatCount > 0 {
		t.repeatCount -= min(t.repeatCount, count)
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
