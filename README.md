# gof

easy way for jsonrpc api, depend on gin

## use

go get -u github.com/atcharles/gof/v2

快速搭建 go jsonrpc 服务器的最佳实践

快速生成 jsonrpc api

内置功能:

- 配置文件加载,支持多格式
- 日志
- 定时任务 cron
- 内存缓存
- resty http 客户端
- 日志文件,按天切割,保存多少天,支持详细配置项
- grace 优雅启动重启,后台任务 finally
- goroutine pool
- 基于 http 的 jsonrpc 2.0,以及 gin 的集成
- 基于 cobra 命令行工具
- xorm 的集成, 扩展工具集,支持单条数据的内存缓存,清理等
- Redis 的订阅发布
- 标准的用户鉴权,token

更多功能有待探索...

## 快速开始

```go
package main

import (
	"github.com/atcharles/gof/v2"
	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/j2rpc"
	"github.com/gin-gonic/gin"
	"github.com/henrylee2cn/goutil"
)

type handler struct {
	API *api `inject:"" j2rpc:""`
}

func (h *handler) Router(*gin.RouterGroup) {}

func (h *handler) J2rpc(j2rpc.RPCServer) {}

type api struct{}

//AddCache ...
func (*api) AddCache(key string, val string) error {
	return gof.App.G2cache.Set(key, []byte(val))
}

//FlushCache ...
func (*api) FlushCache() error {
	return gof.App.G2cache.Reset()
}

//MaxCost ...
func (*api) MaxCost() interface{} {
	return gof.App.G2cache.RistrettoCache().MaxCost()
}

//Get ...
func (*api) Get(key string) (interface{}, error) {
	bts, err := gof.App.G2cache.Get(key)
	if err != nil {
		return nil, err
	}
	return string(bts), nil
}

//Name ...
func (*api) Name() interface{} {
	return goutil.ObjectName(gof.App.G2cache.CacheInstance())
}

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
```

go build -o fast main.go && ./fast start

`request set cache`

```shell
curl -X POST 'http://127.0.0.1:8080/jsonrpc' \
-H 'Content-Type: application/json' \
--data-raw '{"id":1,"method":"api.add_cache","params":["key","value"]}'
```

`request get cache`

```shell
curl -X POST 'http://127.0.0.1:8080/jsonrpc' \
-H 'Content-Type: application/json' \
--data-raw '{"id":1,"method":"api.get","params":["key"]}'
```
