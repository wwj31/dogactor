package tools

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	SecondsPerDay = 24 * 60 * 60
)

var (
	TimeZero       = time.Unix(0, 0)
	TimeOffset     int64
	TimeOffsetPath = "./.timeoffset"
)

func init() {
	bytes, err := os.ReadFile(TimeOffsetPath)
	if err != nil {
		return
	}
	str := string(bytes)
	t, err := strconv.Atoi(str)
	if err != nil {
		return
	}
	TimeOffset = int64(t)
}

func ModifyTimeOffset(add int64) {
	TimeOffset += add
	file, err := os.Create(TimeOffsetPath)
	if err != nil {
		panic(err)
	}
	str := strconv.Itoa(int(TimeOffset))
	_, err = file.Write([]byte(str))
	if err != nil {
		panic(err)
	}
}

func Now() time.Time {
	return time.Now().Add(time.Duration(TimeOffset))
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

func GetNextHour() time.Time {
	return Now().Truncate(time.Hour).Add(time.Hour)
}

func GetNextMinute() time.Time {
	return Now().Truncate(time.Minute).Add(time.Minute)
}

func BeginningOfTheDay(t time.Time) time.Time {
	return t.Truncate(24 * 60 * 60 * time.Second)
}

func EndingOfTheDay(t time.Time) time.Time {
	return t.Truncate(24 * 60 * 60 * time.Second).Add(24 * 60 * 60 * time.Second)
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
