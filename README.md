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


## 缓存锁的生命周期

`g2db.cacheMem.Atomic` 使用 `github.com/moby/locker v1.0.1`：同一缓存键仍然串行执行，不同键允许并行，最后一个等待者完成后删除锁记录。通过 `defer` 保证业务函数发生 panic 时也会解锁。该操作不再把每个历史缓存键永久保存在 `g2db.Locker` 中；公开的业务锁接口保持原有语义。

依赖来自 Docker/Moby 的锁实现，Apache-2.0 许可，仅依赖 Go 标准库，兼容项目现有 Go 版本。选择其稳定数字版本，避免自行实现锁的引用计数或引入哈希分片导致不同键相互阻塞。依赖版本和校验和分别记录在 go.mod、go.sum。官方说明：https://github.com/moby/locker/tree/v1.0.1 。

回归测试覆盖十万个历史键的 GC 后保留内存、跨实例的同键互斥、不同键并行，以及 panic 后可继续操作。维护责任属于 GoF 的 g2db 模块；升级锁库时必须重新运行 `go test -race ./g2db -run TestCacheAtomic`。该修复解决确定的历史键保留问题，不能单独证明某次生产 OOM 的起因。
