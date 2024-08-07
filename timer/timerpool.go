// Copyright 2017-2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// copy from https://github.com/nats-io/nats.go/blob/main/timer.go

package timer

import (
	"sync"
	"time"
)

// global pool of *time.Timer's. can be used by multiple goroutines concurrently.
var globalTimerPool goTimerPool

// timerPool provides GC-able pooling of *time.Timer's.
// can be used by multiple goroutines concurrently.
type goTimerPool struct {
	p sync.Pool
}

// Get returns a timer that completes after the given duration.
func Get(d time.Duration) *time.Timer {
	if t, _ := globalTimerPool.p.Get().(*time.Timer); t != nil {
		t.Reset(d)
		return t
	}

	return time.NewTimer(d)
}

// Put pools the given timer.
//
// There is no need to call t.Stop() before calling Put.
//
// Put will try to stop the timer before pooling. If the
// given timer already expired, Put will read the unreceived
// value if there is one.
func Put(t *time.Timer) {
	if t == nil {
		return
	}

	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	globalTimerPool.p.Put(t)
}
