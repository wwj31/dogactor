package actor

import (
	"time"

	"github.com/wwj31/dogactor/timer"
	"github.com/wwj31/dogactor/tools"
)

type timerScheduler struct {
	*Base

	timerMgr *timer.Manager
	timer    *time.Timer
	nextAt   time.Time
}

func newTimerScheduler(base *Base) *timerScheduler {
	ts := &timerScheduler{
		Base:     base,
		timerMgr: timer.New(),
	}

	return ts
}

// AddTimer default value of timeId is uuid.
// if you need to ensure idempotency of the method,explicit the timeId.
// when timeId is existed the timer will update with endAt and callback and count.
// if times == -1 timer will repack in timer system infinitely util call CancelTimer.
func (s *timerScheduler) AddTimer(timeId string, endAt time.Time, callback func(dt time.Duration), times ...int) string {
	if s.timerMgr == nil {
		return ""
	}

	n := 1
	if len(times) > 0 {
		n = times[0]
	}

	timeId = s.timerMgr.Add(tools.Now(), endAt, callback, n, timeId)

	s.resetTime()
	return timeId
}

func (s *timerScheduler) CancelTimer(timerId string) {
	if s.timerMgr == nil {
		return
	}

	s.timerMgr.Cancel(timerId)
	s.resetTime()
}

// resetTime reset the timer of the timerMgr
func (s *timerScheduler) resetTime(n ...time.Time) {
	if s.timerMgr == nil {
		return
	}

	reset := func(next time.Time) {
		timer.Put(s.timer)

		s.timer = time.AfterFunc(next.Sub(tools.Now()), func() {
			_ = s.Send(s.ID(), func() {
				now := tools.Now()
				if d := s.timerMgr.Update(now); d > 0 {
					s.resetTime(now.Add(d))
				}
			})
		})
		s.nextAt = next
	}

	if len(n) > 0 {
		reset(n[0])
		return
	}

	nextAt := s.timerMgr.NextUpdateAt()
	if nextAt.IsZero() {
		return
	}

	if !nextAt.Equal(s.nextAt) {
		reset(nextAt)
	}
}

func (s *timerScheduler) clear() {
	timer.Put(s.timer)
	s.timerMgr = nil
}
