package g2cmd

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/atcharles/gof/v2/g2db"

	"github.com/atcharles/gof/v2/g2util"
)

// G2cmd ...
type G2cmd struct {
	Logger g2util.LevelLogger `inject:""`
	Config *g2util.Config     `inject:""`
	AbFile *g2util.AbFile     `inject:""`
	Mysql  *g2db.Mysql        `inject:""`

	cmdMap map[string]Process

	startWorkerFunc   func()
	migrateWorkerFunc func()
}

// SetStartWorker ...
func (g *G2cmd) SetStartWorker(fn func()) { g.startWorkerFunc = fn }

// SetMigrateWorker ...
func (g *G2cmd) SetMigrateWorker(fn func()) { g.migrateWorkerFunc = fn }

// CmdMap ...
func (g *G2cmd) CmdMap() map[string]Process { return g.cmdMap }

// Constructor ...
func (g *G2cmd) Constructor() { g.cmdMap = make(map[string]Process) }

// binaryName ...
func (g *G2cmd) binaryName() string {
	return filepath.Join(g.Config.RootPath(), filepath.Base(os.Args[0]))
}

// RegisterCmd ...注册cmd
func (g *G2cmd) RegisterCmd(cmd Process) {
	name := g2util.ValueIndirect(reflect.ValueOf(cmd)).Type().Name()
	g.cmdMap[name] = cmd
}

// Process ...
type Process interface {
	Cmd() *cobra.Command
	SetFlags(cmd *cobra.Command)
	Run(cmd *cobra.Command, args []string)
}
