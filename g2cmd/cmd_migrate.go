package g2cmd

import (
	"log"

	"github.com/spf13/cobra"
)

type migrateCmd struct {
	cmd     *G2cmd
	runFunc func()
	drop    bool
}

func (m *migrateCmd) Cmd() *cobra.Command {
	cmd1 := &cobra.Command{Use: "migrate", Run: m.Run}
	m.SetFlags(cmd1)
	return cmd1
}

func (m *migrateCmd) SetFlags(c *cobra.Command) {
	c.Flags().BoolVarP(&m.drop, "drop", "d", false, "drop database")
}

func (m *migrateCmd) Run(_ *cobra.Command, _ []string) {
	var err error
	if m.drop {
		err = m.cmd.Mysql.DropDatabase()
		if err != nil {
			log.Fatalln(err)
		}
		err = m.cmd.Mysql.Redis.PubDelMemAll()
		if err != nil {
			log.Println(err)
		}
	}
	if m.runFunc != nil {
		m.runFunc()
	}
}
