package store

import (
	"sync"
	"sync/atomic"
	"testing"
)

type testCache struct{ name string }

func (c *testCache) String() string             { return c.name }
func (c *testCache) Set(string, []byte) error   { return nil }
func (c *testCache) Get(string) ([]byte, error) { return nil, ErrNotFound }
func (c *testCache) Delete(string) error        { return nil }
func (c *testCache) Reset() error               { return nil }
func (c *testCache) CacheInstance() ItfCache    { return c }

func TestFactoryCreatesOnlySelectedCacheOnce(t *testing.T) {
	var calls atomic.Int32
	var unusedCalls atomic.Int32
	selected := &testCache{name: t.Name()}
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		delete(factories, selected.name)
		delete(factories, t.Name()+"-unused")
	})
	RegisterFactory(selected.name, func() ItfCache {
		calls.Add(1)
		return selected
	})
	RegisterFactory(t.Name()+"-unused", func() ItfCache {
		unusedCalls.Add(1)
		return &testCache{name: t.Name() + "-unused"}
	})
	if calls.Load() != 0 || unusedCalls.Load() != 0 {
		t.Fatal("注册时不应初始化缓存")
	}
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache, err := GetStore(selected.name)
			if err != nil || cache != selected {
				t.Errorf("并发读取缓存失败: cache=%v err=%v", cache, err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 || unusedCalls.Load() != 0 {
		t.Fatalf("初始化次数错误: selected=%d unused=%d", calls.Load(), unusedCalls.Load())
	}
	RegisterFactory(selected.name, func() ItfCache { t.Fatal("重复注册不应替换缓存"); return nil })
	cache, err := GetStore(selected.name)
	if err != nil || cache != selected {
		t.Fatalf("重复注册改变了实例: cache=%v err=%v", cache, err)
	}
}

func TestRegisterKeepsExistingCacheAndUnknownNameFails(t *testing.T) {
	cache := &testCache{name: t.Name()}
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		delete(ins, cache.name)
	})
	Register(cache)
	RegisterFactory(cache.name, func() ItfCache { t.Fatal("已注册缓存不应调用工厂"); return nil })
	got, err := GetStore(cache.name)
	if err != nil || got != cache {
		t.Fatalf("已注册缓存行为改变: cache=%v err=%v", got, err)
	}
	if _, err := GetStore(t.Name() + "-unknown"); err == nil {
		t.Fatal("未注册的缓存应返回错误")
	}
}
