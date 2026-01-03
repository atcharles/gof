package g2gin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof" //pprof
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/j2rpc"
)

// G2gin ...
type G2gin struct {
	Config   *g2util.Config   `inject:""`
	AbFile   *g2util.AbFile   `inject:""`
	Graceful *g2util.Graceful `inject:""`

	j2opt *j2rpc.Option

	j2Service interface{}
}

// Constructor ...
func (g *G2gin) Constructor() { g.j2opt = j2rpc.SnakeOption }

// SetJ2Service ...
func (g *G2gin) SetJ2Service(j2Service interface{}) { g.j2Service = j2Service }

// Run ...
func (g *G2gin) Run() {
	v := g.Config.Viper().GetStringMap("http_server")
	gin.SetMode(cast.ToString(v["mode"]))
	gin.DisableConsoleColor()
	gin.DefaultWriter = g.AbFile.MustLogIO("web")
	gin.DefaultErrorWriter = gin.DefaultWriter
	eg := gin.New()
	eg.Use(gin.Recovery())
	const pathPrefix = "/"
	g1 := eg.Group(pathPrefix)
	//g1.Use(MiddlewareLimiter(g.Config))
	if apiRoot := strings.TrimPrefix(cast.ToString(v["api_root"]), pathPrefix); len(apiRoot) > 0 {
		g1 = g1.Group(pathPrefix + apiRoot)
	}
	g1.Use(g.copyRequestBody())
	g.useCors(g1)
	g.useJ2rpc(g1)
	srv := &http.Server{
		Addr:              cast.ToString(v["port"]),
		Handler:           eg,
		ReadTimeout:       time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 10,
		IdleTimeout:       time.Second * 30,
	}
	g.Graceful.RegHTTPServer(srv)
	g.profServer()
}

// useJ2rpc ...
func (g *G2gin) useJ2rpc(rg *gin.RouterGroup) {
	jsv := j2rpc.New(g.j2opt)
	jsv.Logger().SetLevel(g2util.ParseLevel(g.Config.Viper().GetString("global.log_level")))
	jsv.Logger().SetOutput(gin.DefaultWriter)
	jsv.Opt().AddBeforeMiddleware(func(c *gin.Context, method string) { c.Set("method", method) }, []string{"^.*$"})
	if g.j2Service != nil {
		jsv.RegisterForApp(g.j2Service)
		if ginRouter, ok := g.j2Service.(ItfGinRouter); ok {
			ginRouter.Router(rg)
			ginRouter.J2rpc(jsv)
		}
	}
	rg.Use(midAddRequestID)
	rg.Any("/jsonrpc", func(c *gin.Context) { jsv.Handler(c, c.Writer, c.Request) })
}

// copyRequestBody ...
func (g *G2gin) copyRequestBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.IsDebugging() {
			if c.Request.Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				c.Set("REQUEST_BODY", bodyBytes)
			}
		}
		c.Next()
	}
}

// UseLogger ...
func (g *G2gin) UseLogger(rg *gin.RouterGroup, writer io.Writer, skipPaths []string) {
	webIO := gin.DefaultWriter
	if writer != nil {
		webIO = writer
	}
	var defaultLogFormatter = func(param gin.LogFormatterParams) string {
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("[GIN]| %v | %s | %3d | %s | %s | %s | %s |",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			cast.ToString(param.Keys["method"]),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		))
		buf.WriteString("\n")
		if len(param.ErrorMessage) > 0 {
			buf.WriteString(param.ErrorMessage)
			buf.WriteString("\n")
		}
		if !gin.IsDebugging() {
			return buf.String()
		}
		buf.WriteString("[Header] ")
		mp := make(map[string]string)
		for k, vs := range param.Request.Header {
			mp[k] = strings.Join(vs, ",")
		}
		buf.WriteString(g2util.JSONDump(mp))
		buf.WriteString("\n")
		buf.WriteString("[Body] ")
		buf.WriteString(cast.ToString(param.Keys["REQUEST_BODY"]))
		buf.WriteString("\n\n")
		return buf.String()
	}
	rg.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: defaultLogFormatter,
		Output:    webIO,
		SkipPaths: skipPaths,
	}))
}

// useCors ...
func (g *G2gin) useCors(rg *gin.RouterGroup) {
	_config := cors.Config{
		AllowAllOrigins: true,
		//AllowOrigins:    []string{},
		//AllowOriginFunc: func(origin string) bool { return true },
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
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
		AllowFiles:             true,
	}
	rg.Use(cors.New(_config))
}

// profServer ...
func (g *G2gin) profServer() {
	port1 := g.Config.Viper().GetString("global.pprof_port")
	if port1 == "" {
		return
	}
	//go get -u github.com/google/pprof
	//需要安装 graphviz
	//pprof -http=:8080 http://127.0.0.1:301/debug/pprof/profile\?seconds\=10
	srv := &http.Server{Addr: port1, Handler: nil}
	g.Graceful.RegHTTPServer(srv)
}
