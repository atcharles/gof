// +build !big_cache

package g2util

import (
	"github.com/atcharles/gof/v2/json"
)

//Set ...
func (g *G2cache) Set(key string, data []byte) (err error) { g.goCache.SetDefault(key, data); return }

//Get ...
func (g *G2cache) Get(key string) (data []byte, err error) {
	val, has := g.goCache.Get(key)
	if !has {
		err = ErrNotFound
		return
	}
	switch d := val.(type) {
	case []byte:
		data = d
	default:
		data, err = json.Marshal(val)
	}
	return
}

//Delete ...
func (g *G2cache) Delete(key string) (err error) { g.goCache.Delete(key); return }

//Reset ...
func (g *G2cache) Reset() (err error) { g.goCache.Flush(); return }
