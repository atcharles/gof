package g2util

import (
	"time"
)

// TimeoutExecFunc ...
func TimeoutExecFunc(fn func(), timeout time.Duration) {
	ch1 := make(chan struct{}, 1)
	go func() {
		fn()
		ch1 <- struct{}{}
	}()
	select {
	case <-time.After(timeout):
		return
	case <-ch1:
		return
	}
}
