package g2db

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/novalagung/gubrak/v2"
	"github.com/unknwon/com"

	"github.com/atcharles/gof/v2/g2cache/store"

	"github.com/atcharles/gof/v2/g2util"
)

var cacheNULL = "null"

type cacheMem struct {
	Logger g2util.LevelLogger `inject:""`
	Cache  store.ItfCache     `inject:""`
}

// RedisSubDelMemAll ...
func (c *cacheMem) RedisSubDelMemAll() RedisSubHandlerFunc {
	return func(_ []byte) {
		if e := c.Cache.Reset(); e != nil {
			c.Logger.Errorf("清空缓存失败:%s", e.Error())
			return
		}
		c.Logger.Debugf("内存缓存已清空")
	}
}

// RedisSubDelCache ...
func (c *cacheMem) RedisSubDelCache() RedisSubHandlerFunc {
	var _fnDel = func(p []byte) (err error) {
		keys := make([]string, 0)
		if err = json.Unmarshal(p, &keys); err != nil {
			return
		}
		for _, key := range keys {
			if e := c.Cache.Delete(key); e != nil {
				c.Logger.Errorf("删除缓存失败:%s; key:", e.Error(), key)
			}
			c.Delete(key)
		}
		return
	}
	return func(payload []byte) {
		if e := _fnDel(payload); e != nil {
			c.Logger.Errorf("删除缓存失败:%s; payload: %s", e.Error(), payload)
			return
		}
		time.AfterFunc(time.Second*2, func() { _ = _fnDel(payload) })
		c.Logger.Debugf("删除缓存: payload: %s", payload)
	}
}

// Constructor New ...
func (c *cacheMem) Constructor() {}

// Atomic ...
func (c *cacheMem) Atomic(key string, fn func()) {
	mu := Locker.Load(fmt.Sprintf("%s:%s", "cacheMem:Atomic", key))
	mu.Lock()
	fn()
	mu.Unlock()
}

// Delete ...
func (c *cacheMem) Delete(key string) {
	c.Atomic(key, func() { _ = c.Cache.Delete(key) })
}

// GetOrStore ...
func (c *cacheMem) GetOrStore(key string, fn func() ([]byte, error)) (data []byte, err error) {
	c.Atomic(key, func() { data, err = c.getOrStore(key, fn) })
	return
}

func (c *cacheMem) getOrStore(key string, fn func() ([]byte, error)) (data []byte, err error) {
	ce := c.Cache
	data, err = ce.Get(key)
	if err != store.ErrNotFound {
		return
	}
	if data, err = fn(); err != nil {
		return
	}
	err = ce.Set(key, data)
	return
}

type cacheBind struct{}

func (*cacheBind) cacheValString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("'%s'", v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	default:
		panic("unSupport kind:" + v.Type().String())
	}
}

// Values ...return bind cache keys
// used for db get
func (*cacheBind) Values(bean interface{}, condition ...interface{}) []string {
	qs := make([]string, 0)
	mp := make(MapString)
	mp.findAllCacheCondition(reflect.ValueOf(bean))
	for k, v := range mp {
		qs = append(qs, fmt.Sprintf("%s=%s", k, v))
	}
	if ci, ok := bean.(ItfCompoundIndex); ok {
		ci1qs := make([]string, 0)
		for _, cia := range ci.CompoundIndexes() {
			v := cia.makeQuery()
			if len(v) > 0 {
				ci1qs = append(ci1qs, v)
			}
		}
		qs = append(qs, ci1qs...)
	}
	if len(condition) > 0 {
		for _, i := range condition {
			switch v := i.(type) {
			case string:
				qs = append(qs, v)
			case []string:
				qs = append(qs, v...)
			case fmt.Stringer:
				qs = append(qs, v.String())
			default:
				qs = append(qs, com.ToStr(v))
			}
		}
	}
	qs = gubrak.From(qs).OrderBy(func(each string) int { return len(each) }, true).Result().([]string)
	return qs
}

// Slice ...
type Slice []interface{}

// Asc ...
func (s Slice) Asc(call interface{}) Slice {
	rs, err := gubrak.From(s).OrderBy(call, true).ResultAndError()
	if err != nil {
		panic(err)
	}
	return rs.(Slice)
}

// MapString ...
type MapString map[string]string

func (mp MapString) findConditionN(val reflect.Value, i int) {
	const xTag, xExt = "xorm", "extends"
	var ks = []string{"pk", "unique"}
	val = g2util.ValueIndirect(val)
	_fn1ins := func(v string, src []string) bool {
		for _, s := range src {
			if strings.Contains(v, s) {
				return true
			}
		}
		return false
	}
	v1f := val.Type().Field(i)
	xv, ok := v1f.Tag.Lookup(xTag)
	if !(ok && !strings.Contains(xv, "-")) {
		return
	}
	xv = strings.ToLower(xv)
	v1v := g2util.ValueIndirect(val.Field(i))
	if xv == xExt {
		mp.findAllCacheCondition(v1v)
		return
	}
	if !_fn1ins(xv, ks) {
		return
	}
	if !(v1v.IsValid() && !v1v.IsZero()) {
		return
	}
	mp[fieldName(v1f.Name)] = new(cacheBind).cacheValString(v1v)
}

func (mp MapString) findAllCacheCondition(val reflect.Value) {
	val = g2util.ValueIndirect(val)
	for i := 0; i < val.NumField(); i++ {
		mp.findConditionN(val, i)
	}
}
