package tools

import (
	"fmt"
	"time"
)

const (
	SecondsPerDay  = 24 * 60 * 60
	SecondsForever = SecondsPerDay * 365 * 10 //10年(int32~(2020+17))
)

var TimeZero = Unix(0, 0)
var TimeZone int

func init() {
	t := Now()
	_, TimeZone = t.Zone()
}

func Now() time.Time {
	return time.Now().UTC()
}

func Unix(sec int64, nsec int64) time.Time {
	return time.Unix(sec, nsec).UTC()
}

func Milliseconds() int64 {
	return Now().UnixNano() / int64(time.Millisecond)
}

func Seconds() int64 {
	return Now().UnixNano() / int64(time.Second)
}

func Date(year int, month time.Month, day, hour, min, sec, nsec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, nsec, time.UTC)
}

func IsSameDay(time1 time.Time, time2 time.Time) bool {
	return time1.Year() == time2.Year() && time1.YearDay() == time2.YearDay()
}

func IsInvalidTime(data time.Time) bool {
	return data.Equal(TimeZero)
}

func TimeFormat(data time.Time) string {
	return data.Format(time.RFC3339Nano)
}

func TimeParse(data string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, data)
	if err != nil {
		return TimeZero
	}
	return t.UTC()
}

func TimeParseFormat(layout, value string) (time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return TimeZero, err
	}
	return t.UTC(), nil
}

func GetNextTime(hour, minute int) time.Time {
	now := Now()

	if hour < 0 || hour > 24 || minute < 0 || minute > 60 {
		fmt.Println("Wrong hour minute", hour, minute)
		return now
	}
	todayTime := BeginningOfTheDay(now).Add(time.Duration(hour) * time.Hour).Add(time.Duration(minute) * time.Minute)
	if todayTime.After(now) {
		return todayTime
	}
	return todayTime.Add(24 * 60 * 60 * time.Second)
}

func GetNextHour() time.Time {
	return Now().Truncate(time.Hour).Add(time.Hour)
}

func GetNextMinute() time.Time {
	return Now().Truncate(time.Minute).Add(time.Minute)
}

func GetTimeWithoutHours(t time.Time) time.Time {
	return t.Truncate(time.Hour).Add(-time.Duration(t.Hour()) * time.Hour)
}

func BeginningOfTheDay(t time.Time) time.Time {
	return t.Truncate(24 * 60 * 60 * time.Second)
}

func MidOfTheDay(t time.Time) time.Time {
	return t.Truncate(24 * 60 * 60 * time.Second).Add(24 * 60 * 60 * time.Second / 2)
}

func EndingOfTheDay(t time.Time) time.Time {
	return t.Truncate(24 * 60 * 60 * time.Second).Add(24 * 60 * 60 * time.Second)
}

func MondayBeginWeek() int64 {
	now := Now()
	offset := int64(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	monday := now.Unix() + offset*SecondsPerDay
	return monday - (monday % SecondsPerDay)
}

func NextMondayBeginWeek() int64 {
	return MondayBeginWeek() + SecondsPerDay*7
}

func DiffDay(t1, t2 time.Time) int32 {
	t1Start := BeginningOfTheDay(t1)
	t2Start := BeginningOfTheDay(t2)
	return int32(t1Start.Sub(t2Start).Seconds() / SecondsPerDay)
}

// 以当天开始时间为初始值 间隔 intervalSeconds触发一次，返回下次触发的时间
func NextIntervalTime(t1 time.Time, intervalSeconds int) time.Time {
	if intervalSeconds <= 0 {
		fmt.Println("wrong1 intervalSeconds")
		return Now()
	}
	if SecondsPerDay%intervalSeconds != 0 {
		fmt.Println("wrong2 intervalSeconds")
	}

	begin := BeginningOfTheDay(t1)
	duration := int(t1.Sub(begin).Seconds())
	next := intervalSeconds * (duration/intervalSeconds + 1)

	return begin.Add(time.Second * time.Duration(next))
}
