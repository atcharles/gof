package gof

import (
	"sync"

	"github.com/atcharles/gof/v2/g2cmd"
	"github.com/atcharles/gof/v2/g2db"
	"github.com/atcharles/gof/v2/g2gin"
	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/j2rpc"
)

//App ...
var App = new(Application)

//Application ...
type Application struct {
	Config *g2util.Config `inject:""`

	Logger   g2util.LevelLogger `inject:""`
	Cron     *g2util.G2cron     `inject:""`
	G2cache  *g2util.G2cache    `inject:""`
	G2resty  *g2util.RestyAgent `inject:""`
	AbFile   *g2util.AbFile     `inject:""`
	Graceful *g2util.Graceful   `inject:""`
	Go       *g2util.GoPool     `inject:""`
	Gin      *g2gin.G2gin       `inject:""`
	G2cmd    *g2cmd.G2cmd       `inject:""`
	Mysql    *g2db.Mysql        `inject:""`
	Token    *g2db.Token        `inject:""`

	populateOnce sync.Once
}

//Populate ...
func (a *Application) Populate() *Application {
	a.populateOnce.Do(func() {
		sysLogger := g2util.NewLevelLogger("[STDOUT]")
		g2util.InjectPopulate(a, sysLogger)
		j2rpc.PopulateConstructor(a)
	})
	return a
}

//Default ...
func (a *Application) Default() *Application {
	a.Populate()
	a.Config.Load("", "conf")
	a.Logger.SetOutput(a.AbFile.MustLogIO("sys"))
	return a
}

//Run ...
func (a *Application) Run() {
	//加载数据库
	a.Mysql.Dial()
	a.Gin.Run()
	a.Graceful.WaitForSignal()
}

//RunServices ...
func (a *Application) RunServices(val interface{}) {
	a.Gin.SetJ2Service(val)
	a.Run()
}

//RunWithCmd ...
func (a *Application) RunWithCmd(fn ...func()) {
	ll := len(fn)
	if ll > 0 {
		a.G2cmd.SetStartWorker(fn[0])
	}
	if ll > 1 {
		a.G2cmd.SetMigrateWorker(fn[1])
	}
	a.G2cmd.Execute()
}

//RunDefault ...
func (a *Application) RunDefault(val interface{}) {
	g2util.InjectPopulate(val, a.Default())
	startFunc := func() {
		a.Gin.SetJ2Service(val)
		a.Run()
	}
	migrateFunc := func() {
		db := a.Mysql
		db.TableRegister(g2util.ObjectTagInstances(val, "migrate")...)
		db.Migrate()
	}
	a.RunWithCmd(startFunc, migrateFunc)
}
