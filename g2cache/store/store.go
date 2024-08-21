package store

import (
	"errors"
	"fmt"
)

// ItfCache ...
type ItfCache interface {
	String() string
	Set(key string, data []byte) (err error)
	Get(key string) (data []byte, err error)
	Delete(key string) (err error)
	Reset() (err error)
	CacheInstance() ItfCache
}

var ins = make(map[string]ItfCache)

// Register ...
func Register(s ItfCache) {
	name := s.String()
	if _, ok := ins[name]; ok {
		//log.Printf("store %s is registered\n", s)
		return
	}
	ins[name] = s
}

func GetStore(names ...string) (ItfCache, error) {
	var name = "ledis"
	if len(names) > 0 {
		name = names[0]
	}
	s, ok := ins[name]
	if !ok {
		return nil, fmt.Errorf("store %s is not registered", name)
	}
	return s, nil
}

// ErrNotFound ...
var ErrNotFound = errors.New("item not found")
