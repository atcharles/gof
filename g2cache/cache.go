package g2cache

import (
	_ "github.com/atcharles/gof/v2/g2cache/cachebig"
	_ "github.com/atcharles/gof/v2/g2cache/cachego"
	_ "github.com/atcharles/gof/v2/g2cache/cacheledis"
	_ "github.com/atcharles/gof/v2/g2cache/cacheristretto"
	"github.com/atcharles/gof/v2/g2cache/store"
)

// Instance ...
type Instance struct{ inc store.ItfCache }

func (i *Instance) Constructor()                      { i.inc, _ = store.GetStore() }
func (i *Instance) String() string                    { return i.inc.String() }
func (i *Instance) Set(key string, data []byte) error { return i.inc.Set(key, data) }
func (i *Instance) Get(key string) ([]byte, error)    { return i.inc.Get(key) }
func (i *Instance) Delete(key string) error           { return i.inc.Delete(key) }
func (i *Instance) Reset() error                      { return i.inc.Reset() }
func (i *Instance) CacheInstance() store.ItfCache     { return i.inc.CacheInstance() }
func (i *Instance) SetInstance(inc store.ItfCache)    { i.inc = inc }
