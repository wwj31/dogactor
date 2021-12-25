package actor

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	tt := time.NewTimer(5 * time.Second)

	b := tt.Reset(2 * time.Second)

	<-tt.C
	println("???????????", b)
}
