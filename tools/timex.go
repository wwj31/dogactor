package tools

import (
	"time"
)

type TimeEx struct {
	time.Time
}

func NewTimeEx(tm ...time.Time) TimeEx {
	if len(tm) == 0 {
		return TimeEx{Time: Now()}
	}
	return TimeEx{Time: tm[0]}
}

func (t TimeEx) DateTime() time.Time {
	return t.Time
}

func (t TimeEx) NextHour() time.Time {
	return t.BeginOfHour().Add(time.Hour)
}

func (t TimeEx) BeginOfMinute() time.Time {
	return t.Truncate(time.Minute)
}

// BeginOfHour 当前小时的起点
func (t TimeEx) BeginOfHour() time.Time {
	return t.Truncate(time.Hour)
}

// BeginOfToday 当日0点
func (t TimeEx) BeginOfToday() time.Time {
	return t.Truncate(24 * time.Hour)
}

// EndOfToday 当日24点
func (t TimeEx) EndOfToday() time.Time {
	return t.BeginOfToday().Add(24 * time.Hour)
}

// StartOfWeek 获取本周周一0点
func (t TimeEx) StartOfWeek() time.Time {
	return t.Truncate(7 * 24 * time.Hour)
}

// EndOfWeek  获取本周周日0点
func (t TimeEx) EndOfWeek() time.Time {
	return t.StartOfWeek().Add(6 * 24 * time.Hour)
}

// StartOfMonthly 获取本月第一天0点
func (t TimeEx) StartOfMonthly() time.Time {
	tm := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	return tm
}

// EndOfMonthly 获取本月最后一天的0点
func (t TimeEx) EndOfMonthly() time.Time {
	tm := t.StartOfMonthly().AddDate(0, 1, -1)
	return tm
}
