package g2util

import (
	"errors"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/patrickmn/go-cache"
)

//G2cache ...
type G2cache struct {
	goCache  *cache.Cache
	bigCache *bigcache.BigCache
}

//BigCache ...
func (g *G2cache) BigCache() *bigcache.BigCache { return g.bigCache }

//GoCache ...
func (g *G2cache) GoCache() *cache.Cache { return g.goCache }

//New ...
func (g *G2cache) New() *G2cache { g.Constructor(); return g }

//Constructor ...
func (g *G2cache) Constructor() {
	g.goCache = cache.New(time.Minute*10, time.Minute*10)
	cfg := bigcache.DefaultConfig(time.Minute * 10)
	cfg.Verbose = false
	g.bigCache, _ = bigcache.NewBigCache(cfg)
}

//ItfCache ...
type ItfCache interface {
	Set(key string, data []byte) (err error)
	Get(key string) (data []byte, err error)
	Delete(key string) (err error)
	Reset() (err error)
}

//ErrNotFound ...
var ErrNotFound = errors.New("item not found")
