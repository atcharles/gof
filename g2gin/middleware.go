package g2gin

import (
	"github.com/didip/tollbooth/v6"
	"github.com/gin-gonic/gin"

	"github.com/atcharles/gof/v2/g2util"
)

// MiddlewareLimiter ...
var MiddlewareLimiter = func(conf *g2util.Config) gin.HandlerFunc {
	requestsPerSecond := conf.Viper().GetFloat64("http_server.limit")
	if requestsPerSecond < 1 {
		requestsPerSecond = float64(5)
	}
	lmt := tollbooth.NewLimiter(requestsPerSecond, nil)
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			lmt.ExecOnLimitReached(c.Writer, c.Request)
			//c.Data(httpError.StatusCode, lmt.GetMessageContentType(), []byte(httpError.Message))
			JSONP(c, httpError.StatusCode, httpError.Message)
			c.Abort()
			return
		}
		c.Next()
	}
}

var midAddRequestID gin.HandlerFunc = func(c *gin.Context) {
	if gin.IsDebugging() {
		c.Header("request-id", g2util.ShortUUID())
	}
	c.Next()
}
