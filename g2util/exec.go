package g2util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ItfExec ...
type ItfExec interface {
	Start() error
	Run() error
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}

// StdExec ...
func StdExec(s string) ItfExec { return NewExecInner(s, os.Stdout) }

// NewExecInner ...
func NewExecInner(s string, out io.Writer) ItfExec {
	cmd := exec.Command("/bin/sh", "-c", s)
	return &execInner{cmd: cmd, out: out}
}

var (
	_ ItfExec = &exec.Cmd{}
	_ ItfExec = &execInner{}
)

type execInner struct {
	cmd *exec.Cmd
	out io.Writer
}

// setOut ...
func (e *execInner) setOut() { e.cmd.Stdout = e.out; e.cmd.Stderr = e.out }

func (e *execInner) Start() error                    { e.setOut(); return e.cmd.Start() }
func (e *execInner) Run() error                      { e.setOut(); return e.cmd.Run() }
func (e *execInner) Output() ([]byte, error)         { return e.cmd.Output() }
func (e *execInner) CombinedOutput() ([]byte, error) { return e.cmd.CombinedOutput() }

//#------------------------------------------------------------------------------------------------------------------#

// FindPidSliceByProcessName get pid list
func FindPidSliceByProcessName(name string) []string {
	str := `ps -ef|grep -v grep|grep '{name}'|awk '{print $2}'|tr -s '\n'`
	p, _ := StdExec(strings.Replace(str, "{name}", name, -1)).Output()
	ps := strings.Split(string(bytes.TrimSpace(p)), "\n")
	return ps
}

// ProcessIsRunning is running
func ProcessIsRunning(name string) bool {
	ps := FindPidSliceByProcessName(name)
	return len(ps) > 0 && len(ps[0]) > 0
}

// KillProcess ...kill process
func KillProcess(name string) (err error) {
	if !ProcessIsRunning(name) {
		return fmt.Errorf("process[%s] is not running", name)
	}
	ps := FindPidSliceByProcessName(name)
	for _, pid := range ps {
		_ = StdExec(fmt.Sprintf("kill %s", pid)).Run()
	}
	return
}
