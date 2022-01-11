package g2util

import (
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/henrylee2cn/goutil"
	"github.com/panjf2000/ants/v2"
)

//GoPool ...goroutine pool
type GoPool struct {
	LevelLogger LevelLogger `inject:""`
	Grace       *Graceful   `inject:""`

	wait *sync.WaitGroup
	pool *ants.Pool
}

//Constructor ...
func (g *GoPool) Constructor() {
	g.wait = new(sync.WaitGroup)
	g.pool, _ = ants.NewPool(runtime.NumCPU()*100, ants.WithOptions(ants.Options{
		PanicHandler: func(p interface{}) {
			panicData := goutil.PanicTrace(4)
			g.LevelLogger.Printf("[Goroutine] worker exits from a panic: [%v] \n%s\n\n", p, panicData)
		},
	}))
	g.Grace.RegProcessor(g)
}

//AfterShutdown ...
func (g *GoPool) AfterShutdown() { g.wait.Wait(); g.pool.Release() }

//Pool ...
func (g *GoPool) Pool() *ants.Pool { return g.pool }

type goFunc func() (err error)

//goFunc ...
func (g *GoPool) goFuncDo(fn goFunc) {
	if e := fn(); e != nil {
		//g.LevelLogger.Warnf("Goroutine worker exits with a error: %s\n\n", e.Error())
		g.LevelLogger.Errorf("[Goroutine] %s\n%s\n", e.Error(), debug.Stack())
	}
}

//Submit ...
func (g *GoPool) Submit(fn goFunc) {
	g.wait.Add(1)
	_ = g.pool.Submit(func() { defer g.wait.Done(); g.goFuncDo(fn) })
}

//Go ... 直接执行,不加入wait队列
func (g *GoPool) Go(fn goFunc) { _ = g.pool.Submit(func() { g.goFuncDo(fn) }) }
