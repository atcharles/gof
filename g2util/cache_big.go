// +build big_cache

package g2util

func (g *G2cache) Set(key string, data []byte) (err error) { return g.bigCache.Set(key, data) }

func (g *G2cache) Get(key string) (data []byte, err error) { return g.bigCache.Get(key) }

func (g *G2cache) Delete(key string) (err error) { return g.bigCache.Delete(key) }

func (g *G2cache) Reset() (err error) { return g.bigCache.Reset() }
