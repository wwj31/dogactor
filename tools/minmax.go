package tools

import (
	"time"
)

type Comparable interface {
	int | int8 | int16 | int32 | int64 | float64 | time.Duration | string
}

func Max[Value Comparable](a, b Value) Value {
	if a > b {
		return a
	}
	return b
}

func Min[Value Comparable](a, b Value) Value {
	if a < b {
		return a
	}
	return b
}

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}
