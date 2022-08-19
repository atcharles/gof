package g2cmd

import (
	"log"

	"github.com/spf13/cobra"
)

type rootCmd struct {
	cmd *G2cmd

	runFunc    func(cmd *cobra.Command)
	cleanCache bool
}

// SetRunFunc ...
func (r *rootCmd) SetRunFunc(runFunc func(cmd *cobra.Command)) { r.runFunc = runFunc }

// FlagsGetBoolFunc ...
func FlagsGetBoolFunc(cmd *cobra.Command, name string, fn func()) {
	ok, err := cmd.Flags().GetBool(name)
	if err != nil {
		log.Fatalln(err)
	}
	if ok {
		fn()
	}
}

func (r *rootCmd) Cmd() *cobra.Command {
	cmd1 := &cobra.Command{Use: r.cmd.binaryName(), Run: r.Run}
	r.SetFlags(cmd1)
	return cmd1
}

func (r *rootCmd) SetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP("sd", "s", false, "set daemon")
	cmd.PersistentFlags().BoolVarP(&r.cleanCache, "clean-cache", "c", false, "clean cache")
	cmd.Flags().BoolP("version", "v", false, "")
}

func (r *rootCmd) Run(cmd *cobra.Command, _ []string) {
	FlagsGetBoolFunc(cmd, "version", showVersion)
	if r.cleanCache {
		err := r.cmd.Mysql.Redis.PubDelMemAll()
		if err != nil {
			log.Println(err)
		}
		log.Printf("清理内存缓存命令已发出\n")
	}
	if r.runFunc != nil {
		r.runFunc(cmd)
	}
}
