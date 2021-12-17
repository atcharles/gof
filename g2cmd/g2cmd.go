package g2cmd

import (
	"log"

	"github.com/atcharles/gof/v2/g2util"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (g *G2cmd) Execute() {
	BuildInfoInstanceAdd(g2util.Map{"appName": g.Config.Viper().GetString("name")})
	g.RegisterCmd(&rootCmd{cmd: g})
	g.RegisterCmd(&startCmd{cmd: g, worker: g.startWorkerFunc})
	g.RegisterCmd(&stopCmd{cmd: g})
	g.RegisterCmd(&restartCmd{cmd: g})
	g.RegisterCmd(&migrateCmd{runFunc: g.migrateWorkerFunc})

	root := g.cmdMap["rootCmd"].Cmd()
	for _, process := range g.cmdMap {
		root.AddCommand(process.Cmd())
	}
	if e := root.Execute(); e != nil {
		log.Fatalln(e)
	}
}
