package cachego

import (
	"encoding/json"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/atcharles/gof/v2/g2cache/store"
)

// GoCache ...
type GoCache struct {
	inc *cache.Cache
}

func (g *GoCache) Instance() *cache.Cache { return g.inc }

// New ...
func (*GoCache) New() *GoCache {
	inc := new(GoCache)
	inc.Constructor()
	return inc
}

func (g *GoCache) Constructor()   { g.inc = cache.New(time.Minute*10, time.Minute*10) }
func (g *GoCache) String() string { return "gocache" }
func (g *GoCache) Set(key string, data []byte) (err error) {
	g.inc.Set(key, data, cache.NoExpiration)
	return
}
func (g *GoCache) Get(key string) (data []byte, err error) {
	val, has := g.inc.Get(key)
	if !has {
		err = store.ErrNotFound
		return
	}
	switch d := val.(type) {
	case string:
		data = []byte(d)
	case []byte:
		data = d
	default:
		data, err = json.Marshal(d)
	}
	return
}
func (g *GoCache) Delete(key string) (err error) { g.inc.Delete(key); return }
func (g *GoCache) Reset() (err error)            { g.inc.Flush(); return }
func (g *GoCache) CacheInstance() store.ItfCache { return g }

func init() {
	store.Register(new(GoCache).New())
}
