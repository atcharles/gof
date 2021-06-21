package app

import (
	"github.com/gin-gonic/gin"

	"github.com/atcharles/gof/v2"
	"github.com/atcharles/gof/v2/j2rpc"
)

type service struct {
	AA1 *api1 `inject:""`
}

func (s *service) Router(g *gin.RouterGroup) {}

func (s *service) J2rpc(jsv j2rpc.RPCServer) {}

type api1 struct{}

func (a *api1) Constructor() {}

//SayHi ...
func (a *api1) SayHi() string { return "Hi" }

//Echo ...
func (a *api1) Echo(data interface{}) interface{} { return data }

//Run ...
func Run() { gof.App.RunServices(new(service)) }
