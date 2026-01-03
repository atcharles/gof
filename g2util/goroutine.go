package g2util

import (
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andeya/goutil"
	"github.com/panjf2000/ants/v2"
)

var goPoolSize = runtime.NumCPU() * 1000

// GoPool ...goroutine pool
type GoPool struct {
	LevelLogger LevelLogger `inject:""`
	Grace       *Graceful   `inject:""`

	wait *sync.WaitGroup
	pool *ants.Pool

	mu   sync.RWMutex
	size atomic.Int32
}

// Constructor ...
func (g *GoPool) Constructor() {
	g.wait = new(sync.WaitGroup)
	g.pool, _ = ants.NewPool(goPoolSize, ants.WithOptions(ants.Options{
		PanicHandler: func(p interface{}) {
			panicData := goutil.PanicTrace(4)
			g.LevelLogger.Printf("[Goroutine] worker exits from a panic: [%v] \n%s\n\n", p, panicData)
		},
	}))
	g.Grace.RegProcessor(g)
}

// AfterShutdown ...
func (g *GoPool) AfterShutdown() {
	TimeoutExecFunc(g.wait.Wait, time.Second*30)
	g.pool.Release()
}

// Pool ...
func (g *GoPool) Pool() *ants.Pool { return g.pool }

type goFunc func() (err error)

// goFunc ...
func (g *GoPool) goFuncDo(fn goFunc) {
	if e := fn(); e != nil {
		//g.LevelLogger.Warnf("Goroutine worker exits with a error: %s\n\n", e.Error())
		g.LevelLogger.Errorf("[Goroutine] %s\n%+v\n%s\n", e.Error(), e, debug.Stack())
	}
}

// Submit ...
//func (g *GoPool) Submit(fn goFunc) {
//	g.wait.Add(1)
//	_ = g.pool.Submit(func() {
//		defer g.wait.Done()
//		g.goFuncDo(fn)
//	})
//}

// Go ... 直接执行,不加入wait队列
func (g *GoPool) Go(fn goFunc) {
	g.size.Add(1)
	g.mu.RLock()
	if g.size.Load() >= int32(goPoolSize-10) {
		g.LevelLogger.Errorf("[Goroutine] goroutine pool is full, size: %d", g.size.Load())
	}
	g.mu.RUnlock()
	_ = g.pool.Submit(func() {
		g.goFuncDo(fn)
		g.size.Add(-1)
	})
}
