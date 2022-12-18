package cacheledis

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/andeya/goutil"
	"github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
	"github.com/pkg/errors"

	"github.com/atcharles/gof/v2/g2cache/store"
)

// LeDis ...
type LeDis struct {
	cfg *config.Config
	inc *ledis.Ledis
	db  *ledis.DB

	once sync.Once
}

func (l *LeDis) SetCfg(cfg *config.Config) { l.cfg = cfg }
func (l *LeDis) Config() *config.Config    { return l.cfg }
func (l *LeDis) SetInc(inc *ledis.Ledis)   { l.inc = inc }
func (l *LeDis) Instance() *ledis.Ledis {
	l.once.Do(func() {
		err := l.SetInstance(nil)
		if err != nil {
			log.Fatalf("load ledis: %s\n", err.Error())
		}
	})
	return l.inc
}
func (l *LeDis) SetDB(db *ledis.DB) { l.db = db }
func (l *LeDis) DB() *ledis.DB      { l.Instance(); return l.db }

// StoreRoot ...
var StoreRoot = goutil.SelfDir()

// SetInstance ...
func (l *LeDis) SetInstance(cfg *config.Config) (err error) {
	if cfg == nil {
		cfg = config.NewConfigDefault()
		cfg.DataDir = filepath.Join(StoreRoot, "._ledisdata")
		//cfg.LMDB.MapSize = 2 * config.GB
	}
	l.cfg = cfg
	l.inc, err = ledis.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "open")
	}
	l.db, err = l.inc.Select(0)
	if err != nil {
		return errors.Wrap(err, "select")
	}
	return
}

// New ...
func (*LeDis) New() *LeDis {
	ln := new(LeDis)
	ln.Constructor()
	return ln
}

// Constructor ...
func (l *LeDis) Constructor() {}

func (l *LeDis) String() string { return "ledis" }

// Set ...
func (l *LeDis) Set(key string, data []byte) (err error) { return l.DB().Set([]byte(key), data) }

// Get ...
func (l *LeDis) Get(key string) (data []byte, err error) {
	data, err = l.DB().Get([]byte(key))
	if err != nil {
		return
	}
	if data == nil {
		err = store.ErrNotFound
		return
	}
	return
}

// Delete ...
func (l *LeDis) Delete(key string) (err error) {
	_, err = l.DB().Del([]byte(key))
	return
}

// Reset ...
func (l *LeDis) Reset() (err error) {
	_, err = l.DB().FlushAll()
	return
}

// CacheInstance ...
func (l *LeDis) CacheInstance() store.ItfCache { return l }

func init() {
	store.Register(new(LeDis).New())
}
