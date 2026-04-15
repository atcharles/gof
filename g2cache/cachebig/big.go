package cachebig

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"

	"github.com/atcharles/gof/v2/g2cache/store"
)

func init() {
	store.Register(new(BigCache).New())
}

// BigCache ...
type BigCache struct {
	inc *bigcache.BigCache
}

func (b *BigCache) CacheInstance() store.ItfCache { return b }

// Constructor ...
func (b *BigCache) Constructor() {
	var err error
	cfg := bigcache.DefaultConfig(time.Minute * 10)
	cfg.Verbose = false
	b.inc, err = bigcache.New(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
}

func (b *BigCache) Delete(key string) (err error) { return b.inc.Delete(key) }

func (b *BigCache) Get(key string) (data []byte, err error) {
	data, err = b.inc.Get(key)
	if err != nil {
		err = store.ErrNotFound
		return
	}
	return
}

func (b *BigCache) Instance() *bigcache.BigCache { return b.inc }

// New ...
func (*BigCache) New() *BigCache {
	inc := new(BigCache)
	inc.Constructor()
	return inc
}

func (b *BigCache) Reset() (err error) { return b.inc.Reset() }

func (b *BigCache) Set(key string, data []byte) (err error) { return b.inc.Set(key, data) }

func (b *BigCache) String() string { return "big" }
