package cacheristretto

import (
	"encoding/json"

	"github.com/dgraph-io/ristretto"

	"github.com/atcharles/gof/v2/g2cache/store"
)

// Ristretto ...
type Ristretto struct {
	inc *ristretto.Cache
}

func (r *Ristretto) Instance() *ristretto.Cache { return r.inc }

// New ...
func (*Ristretto) New() *Ristretto {
	inc := new(Ristretto)
	inc.Constructor()
	return inc
}

func (r *Ristretto) Constructor() {
	var err error
	r.inc, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,           // number of keys to track frequency of (10M).
		MaxCost:     (1 << 30) * 1, // maximum cost of cache (1GB).
		BufferItems: 64,            // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
}

func (r *Ristretto) String() string { return "ristretto" }

func (r *Ristretto) Set(key string, data []byte) (err error) {
	r.inc.SetWithTTL(key, data, int64(len(data)), 0)
	return
}

func (r *Ristretto) Get(key string) (data []byte, err error) {
	val, has := r.inc.Get(key)
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

func (r *Ristretto) Delete(key string) (err error) { r.inc.Del(key); return }

func (r *Ristretto) Reset() (err error) { r.inc.Clear(); return }

func (r *Ristretto) CacheInstance() store.ItfCache { return r }

func init() {
	store.Register(new(Ristretto).New())
}
