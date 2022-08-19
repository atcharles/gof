package g2util

import (
	"time"
)

// Ticker ...
func Ticker(d time.Duration, fn func()) {
	tk := time.NewTicker(d)
	defer tk.Stop()
	for {
		<-tk.C
		fn()
	}
}
