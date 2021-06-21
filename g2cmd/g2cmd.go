package g2cmd

import (
	"log"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (g *G2cmd) Execute() {
	root := g.cmdMap["rootCmd"].Cmd()
	for _, process := range g.cmdMap {
		root.AddCommand(process.Cmd())
	}
	if e := root.Execute(); e != nil {
		log.Fatalln(e)
	}
}
