package store

import (
	"errors"
	"fmt"
	"sync"
)

// ErrNotFound ...
var ErrNotFound = errors.New("item not found")
var ins = make(map[string]ItfCache)
var factories = make(map[string]*cacheFactory)
var registryMu sync.RWMutex

type cacheFactory struct {
	once   sync.Once
	create func() ItfCache
	cache  ItfCache
}

// ItfCache ...
type ItfCache interface {
	String() string
	Set(key string, data []byte) (err error)
	Get(key string) (data []byte, err error)
	Delete(key string) (err error)
	Reset() (err error)
	CacheInstance() ItfCache
}

func GetStore(names ...string) (ItfCache, error) {
	var name = "ledis"
	if len(names) > 0 {
		name = names[0]
	}
	registryMu.RLock()
	s, ok := ins[name]
	factory := factories[name]
	registryMu.RUnlock()
	if ok {
		return s, nil
	}
	if factory == nil {
		return nil, fmt.Errorf("store %s is not registered", name)
	}
	factory.once.Do(func() { factory.cache = factory.create() })
	return factory.cache, nil
}

// RegisterFactory 按需初始化缓存，避免未选用的缓存提前分配内存。
func RegisterFactory(name string, create func() ItfCache) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := ins[name]; ok {
		return
	}
	if _, ok := factories[name]; ok {
		return
	}
	factories[name] = &cacheFactory{create: create}
}

// Register ...
func Register(s ItfCache) {
	name := s.String()
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := factories[name]; ok {
		return
	}
	if _, ok := ins[name]; ok {
		//log.Printf("store %s is registered\n", s)
		return
	}
	ins[name] = s
}
