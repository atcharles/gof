package g2cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/atcharles/gof/v2/g2util"
)

type stopCmd struct {
	cmd *G2cmd
}

func (s *stopCmd) Cmd() *cobra.Command {
	cmd1 := &cobra.Command{Use: "stop", Run: s.Run}
	s.SetFlags(cmd1)
	return cmd1
}

func (s *stopCmd) SetFlags(*cobra.Command) {}

func (s *stopCmd) Run(*cobra.Command, []string) {
	log.Println("Stop server -> Waiting ...")
	binaryName := s.cmd.binaryName()
	proName := fmt.Sprintf("%s -s start", filepath.Base(binaryName))
	if err := g2util.KillProcess(proName); err != nil {
		log.Println(err)
		return
	}
	ch1 := make(chan struct{})
	go func() {
		tk := time.NewTicker(time.Millisecond * 500)
		defer tk.Stop()
		for {
			<-tk.C
			if !g2util.ProcessIsRunning(proName) {
				ch1 <- struct{}{}
				return
			}
		}
	}()
	<-ch1
	close(ch1)
	log.Println("Server Stopped OK!")
}
