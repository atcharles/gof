package g2cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/atcharles/gof/v2/g2util"
)

type startCmd struct {
	cmd *G2cmd

	daemon bool
	worker func()
}

//SetWorker 设置启动程序
func (s *startCmd) SetWorker(worker func()) { s.worker = worker }

func (s *startCmd) Cmd() *cobra.Command {
	cmd1 := &cobra.Command{Use: "start", Run: s.Run}
	s.SetFlags(cmd1)
	return cmd1
}

func (s *startCmd) SetFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&s.daemon, "daemon", "d", false, "")
}

func (s *startCmd) Run(*cobra.Command, []string) {
	if !s.daemon {
		//log.Printf("The pid at which the program is currently running is:👉 %d\n", os.Getpid())
		s.cmd.Logger.Println(fmt.Sprintf("Program's PID:👉 %d", os.Getpid()))
		s.worker()
		return
	}
	binaryName := s.cmd.binaryName()
	proName := fmt.Sprintf("%s -s start", filepath.Base(binaryName))
	if g2util.ProcessIsRunning(proName) {
		log.Println("程序已运行,如需重启,请运行 restart 命令")
		return
	}
	cmdString := fmt.Sprintf("%s -s start", binaryName)
	output := s.cmd.AbFile.MustLogIO("output")
	_ = output.File().Truncate(0)
	_ = g2util.NewExecInner(cmdString, output.File()).Start()
	log.Println("Server Started OK!")
}
