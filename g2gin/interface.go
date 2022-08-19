package g2gin

import (
	"github.com/gin-gonic/gin"

	"github.com/atcharles/gof/v2/j2rpc"
)

// ItfGinRouter ...gin router interface
type ItfGinRouter interface {
	Router(g *gin.RouterGroup)
	J2rpc(jsv j2rpc.RPCServer)
}
