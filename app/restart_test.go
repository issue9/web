// SPDX-License-Identifier: MIT

package app

import (
	"net/http"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/server"
)

func TestSignalHUP(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	a := assert.New(t, false)

	exit := make(chan struct{}, 10)
	cmd := &CLIOf[empty]{
		Name:           "test",
		Version:        "1.0.0",
		ConfigFilename: "web.yaml",
		ServeActions:   []string{"serve"},
		Init: func(s *server.Server, user *empty, act string) error {
			return nil
		},
	}
	SignalHUP(cmd)

	go func() {
		a.ErrorIs(cmd.Exec([]string{"app", "-f=./testdata", "-a=serve"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(500 * time.Millisecond) // 等待 go func 启动完成

	p, err := os.FindProcess(os.Getpid())
	a.NotError(err).NotNil(p)

	// hup1
	t1 := cmd.app.srv.Uptime()
	a.NotError(p.Signal(syscall.SIGHUP))
	time.Sleep(500 * time.Millisecond) // 此值要大于 AppOf.ShutdownTimeout
	t2 := cmd.app.srv.Uptime()
	a.True(t2.After(t1))

	// hup2
	a.NotError(p.Signal(syscall.SIGHUP))
	time.Sleep(500 * time.Millisecond) // 此值要大于 AppOf.ShutdownTimeout
	t3 := cmd.app.srv.Uptime()
	a.True(t3.After(t2))

	cmd.app.srv.Close(0)
	<-exit
}
