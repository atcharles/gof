package g2gin

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/j2rpc"
)

//G2gin ...
type G2gin struct {
	Config   *g2util.Config   `inject:""`
	AbFile   *g2util.AbFile   `inject:""`
	Graceful *g2util.Graceful `inject:""`

	j2opt *j2rpc.Option

	j2Service interface{}
}

//Constructor ...
func (g *G2gin) Constructor() { g.j2opt = j2rpc.SnakeOption }

//SetJ2Service ...
func (g *G2gin) SetJ2Service(j2Service interface{}) { g.j2Service = j2Service }

//Run ...
func (g *G2gin) Run() {
	v := g.Config.Viper().GetStringMap("http_server")

	gin.SetMode(cast.ToString(v["mode"]))
	gin.DisableConsoleColor()
	eg := gin.New()
	eg.AppEngine = true
	err := eg.SetTrustedProxies([]string{"0.0.0.0/0"})
	if err != nil {
		log.Fatalln(err)
	}

	g1 := eg.Group("/").Group(cast.ToString(v["api_root"]))
	g.useLogger(g1)
	g.useCors(g1)
	g.useJ2rpc(g1)

	g1.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	srv := &http.Server{
		Addr:              cast.ToString(v["port"]),
		Handler:           eg,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 3,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 30,
	}
	g.Graceful.RegHTTPServer(srv)
}

//useJ2rpc ...
func (g *G2gin) useJ2rpc(rg *gin.RouterGroup) {
	jsv := j2rpc.New(g.j2opt)
	jsv.Logger().SetOutput(gin.DefaultWriter)

	jsv.Opt().AddBeforeMiddleware([]string{"^.*$"}, []string{}, func(c *gin.Context, method string) { c.Set("method", method) })

	if g.j2Service != nil {
		jsv.RegisterForApp(g.j2Service)

		if ginRouter, ok := g.j2Service.(ItfGinRouter); ok {
			ginRouter.Router(rg)
			ginRouter.J2rpc(jsv)
		}
	}
	rg.Any("/jsonrpc", func(c *gin.Context) { jsv.Handler(c, c.Writer, c.Request) })
}

//useLogger ...
func (g *G2gin) useLogger(rg *gin.RouterGroup) {
	webIO := g.AbFile.MustLogIO("web")
	gin.DefaultWriter = webIO
	gin.DefaultErrorWriter = webIO

	var defaultLogFormatter = func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		return fmt.Sprintf("[GIN] %v | %s\n|%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Keys["method"],
			statusColor,
			param.StatusCode,
			resetColor,
			param.Latency,
			param.ClientIP,
			methodColor,
			param.Method,
			resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}

	rg.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: defaultLogFormatter,
		Output:    webIO,
		SkipPaths: nil,
	}))
	rg.Use(gin.RecoveryWithWriter(webIO))
}

//useCors ...
func (g *G2gin) useCors(rg *gin.RouterGroup) {
	_config := cors.Config{
		AllowAllOrigins: false,
		AllowOrigins:    nil,
		AllowOriginFunc: func(origin string) bool { return true },
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type",
			"Accept-Encoding", "Authorization", "X-Request-ID",
			"X-Token", "X-Server", "X-Requested-With",
			"Token",
		},
		AllowCredentials:       true,
		ExposeHeaders:          []string{"X-Token", "X-Server"},
		MaxAge:                 12 * time.Hour,
		AllowWildcard:          true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
		AllowFiles:             false,
	}
	rg.Use(cors.New(_config))
}
