package gosrc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var Gotool = filepath.Join(runtime.GOROOT(), "bin/go")

type Report struct {
	Command *exec.Cmd
	Started time.Time
	Ended   time.Time
	Error   error
	Stdout  bytes.Buffer
	Stderr  bytes.Buffer
}

func (r *Report) Start() error {
	r.Started = time.Now()
	r.Error = r.Command.Start()
	return r.Error
}
func (r *Report) Wait() error {
	if r.Error != nil {
		return r.Error
	}
	r.Error = r.Command.Wait()
	r.Ended = time.Now()
	return r.Error
}
func (r *Report) WaitTimeout(after time.Duration) error {
	if r.Error != nil {
		return r.Error
	}
	done := make(chan bool)
	go func() {
		r.Wait()
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(after):
		r.Command.Process.Kill()
		r.Error = fmt.Errorf("timeout")
	}
	return r.Error
}

func Install(pkg *Pkg) *Report {
	r := gocmd("install", pkg.Path)
	r.Start()
	if r.Error == nil {
		r.WaitTimeout(time.Second)
	}
	return r
}

func Test(pkg *Pkg) *Report {
	r := gocmd("test", "-v", pkg.Path)
	r.Start()
	if r.Error == nil {
		r.WaitTimeout(time.Second)
	}
	return r
}

func gocmd(args ...string) *Report {
	r := &Report{}
	r.Command = &exec.Cmd{
		Path:   Gotool,
		Args:   append([]string{"go"}, args...),
		Dir:    os.TempDir(),
		Stdout: &r.Stdout,
		Stderr: &r.Stderr,
	}
	return r
}
