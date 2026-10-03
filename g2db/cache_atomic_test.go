package g2db

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestCacheAtomicAllowsDifferentKeys(t *testing.T) {
	c := new(cacheMem)
	held, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		c.Atomic("held-regression-key", func() { close(held); <-release })
	}()
	<-held
	defer close(release)
	go func() {
		c.Atomic("independent-regression-key", func() {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("不同键被不必要地串行阻塞")
	}
}

func TestCacheAtomicReleasesHistoricalKeys(t *testing.T) {
	c := new(cacheMem)
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := range 100000 {
		c.Atomic(fmt.Sprintf("expired-issue-memory-regression:%d", i), func() {})
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(c)
	retained := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	t.Logf("100000 个已完成的不同键操作，GC 后保留增量：%d 字节", retained)
	if retained > 4<<20 {
		t.Fatalf("历史键仍长期保留内存：%d 字节", retained)
	}
}

func TestCacheAtomicReleasesOnPanic(t *testing.T) {
	c := new(cacheMem)
	func() {
		defer func() { _ = recover() }()
		c.Atomic("panic-regression-key", func() { panic("模拟业务异常") })
	}()
	done := make(chan struct{})
	go func() {
		c.Atomic("panic-regression-key", func() {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("业务异常后相同键被永久锁住")
	}
}

func TestCacheAtomicSerializesAcrossInstances(t *testing.T) {
	var first, second cacheMem
	var wg sync.WaitGroup
	value := 0
	for i := range 32 {
		c := &first
		if i%2 != 0 {
			c = &second
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				c.Atomic("shared-regression-key", func() { value++ })
			}
		}()
	}
	wg.Wait()
	if value != 32000 {
		t.Fatalf("相同键的并发操作丢失更新：%d", value)
	}
}
