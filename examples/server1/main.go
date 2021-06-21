package main

import (
	"github.com/atcharles/gof/v2"
	"github.com/atcharles/gof/v2/g2util"
)

type service struct {
	AA1 *api1 `inject:""`
}

type api1 struct{}

func (a *api1) Constructor() {}

//SayHi ...
func (a *api1) SayHi() string { return "Hi" }

func main() {
	app := new(gof.Application).Default()
	srv := new(service)
	g2util.InjectPopulate(srv)
	app.Gin.SetJ2Service(srv)
	app.Run()
}
