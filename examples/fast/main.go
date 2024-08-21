package main

import (
	"github.com/andeya/goutil"
	"github.com/gin-gonic/gin"

	"github.com/atcharles/gof/v2"
	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/j2rpc"
)

type handler struct {
	API *api `inject:"" j2rpc:""`
}

func (h *handler) Router(*gin.RouterGroup) {}

func (h *handler) J2rpc(j2rpc.RPCServer) {}

type api struct{}

// AddCache ...
func (*api) AddCache(key string, val string) error {
	return gof.App.G2cache.Set(key, []byte(val))
}

// FlushCache ...
func (*api) FlushCache() error {
	return gof.App.G2cache.Reset()
}

// Get ...
func (*api) Get(key string) (interface{}, error) {
	bts, err := gof.App.G2cache.Get(key)
	if err != nil {
		return nil, err
	}
	return string(bts), nil
}

// Name ...
func (*api) Name() interface{} {
	return goutil.ObjectName(gof.App.G2cache.CacheInstance())
}

// String ...
func (*api) String() string { return gof.App.G2cache.String() }

func main() {
	a, val := gof.App, new(handler)
	g2util.InjectPopulate(val, a.Default())
	startFunc := func() {
		a.Gin.SetJ2Service(val)
		a.Gin.Run()
		a.Graceful.WaitForSignal()
	}
	migrateFunc := func() {}
	a.RunWithCmd(startFunc, migrateFunc)
}
