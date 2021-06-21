package g2cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/atcharles/gof/v2/g2util"
)

type restartCmd struct {
	cmd *G2cmd
}

func (s *restartCmd) Cmd() *cobra.Command {
	cmd1 := &cobra.Command{Use: "restart", Run: s.Run}
	s.SetFlags(cmd1)
	return cmd1
}

func (s *restartCmd) SetFlags(*cobra.Command) {}

func (s *restartCmd) Run(*cobra.Command, []string) {
	binaryName := s.cmd.binaryName()
	_ = g2util.StdExec(fmt.Sprintf("%s stop", binaryName)).Run()
	_ = g2util.StdExec(fmt.Sprintf("%s start -d", binaryName)).Run()
}
