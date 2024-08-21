package g2db

import (
	"sync"
)

// Locker ...
var Locker = new(lock)

type lock struct{ mp sync.Map }

// Load ...
func (l *lock) Load(key string) *sync.RWMutex {
	actual, _ := l.mp.LoadOrStore(key, new(sync.RWMutex))
	return actual.(*sync.RWMutex)
}
