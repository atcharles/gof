package main

import (
	"context"
	"log"
	"net/http"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"

	"github.com/atcharles/gof/v2/j2rpc"
)

type aa struct{}

//ExcludeMethod ...添加排除api函数
func (a *aa) ExcludeMethod() []string { return []string{"API1"} }

//API1 ...
func (a *aa) API1() (data interface{}) { return "hello world" }

//API2 ...
func (a *aa) API2(c context.Context) (data interface{}) { return reflect.TypeOf(c).String() }

//API3 ...
func (a *aa) API3(val string) (data interface{}) { return val }

func main() {
	opt := j2rpc.SnakeOption
	opt.AddBeforeMiddleware(
		[]string{`aa.api1`, `^aa\.\S+[1]$`},
		[]string{`^aa\.open.*$`},
		func(c context.Context, method string, w http.ResponseWriter, r *http.Request) (err error) {
			spew.Dump(reflect.TypeOf(c).Elem().Name())
			j2rpc.AbortWriteHeader(w, 401)
			return j2rpc.NewError(j2rpc.ErrServer, method)
		},
	)

	rpc1server := j2rpc.New(opt)
	rpc1server.Register(new(aa))

	s := http.NewServeMux()
	s.Handle("/jsonrpc", rpc1server)
	go func() { log.Fatalln(http.ListenAndServe(":301", s)) }()

	s1 := gin.Default()
	s1.Any("/jsonrpc", func(c *gin.Context) { rpc1server.Handler(c, c.Writer, c.Request) })
	log.Fatalln(s1.Run(":300"))
}
