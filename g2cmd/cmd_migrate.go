package g2cmd

import (
	"github.com/spf13/cobra"
)

type migrateCmd struct {
	runFunc func()
}

func (m *migrateCmd) Cmd() *cobra.Command {
	cmd1 := &cobra.Command{Use: "migrate", Run: m.Run}
	m.SetFlags(cmd1)
	return cmd1
}

func (m *migrateCmd) SetFlags(_ *cobra.Command) {}

func (m *migrateCmd) Run(_ *cobra.Command, _ []string) {
	if m.runFunc != nil {
		m.runFunc()
	}
}
