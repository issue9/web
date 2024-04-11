// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"net/http"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

var (
	_ App = &app{}
	_ App = &CLI[empty]{}
)

func TestSignalHUP(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	a := assert.New(t, false)

	exit := make(chan struct{}, 10)
	cmd := &CLI[empty]{
		Name:           "test",
		Version:        "1.0.0",
		ConfigDir:      ".",
		ConfigFilename: "web.yaml",
		ServeActions:   []string{"serve"},
		NewServer: func(name, ver string, opt *server.Options, _ *empty, _ string) (web.Server, error) {
			return server.New(name, ver, opt)
		},
	}
	SignalHUP(cmd)

	go func() {
		a.ErrorIs(cmd.exec([]string{"app", "-a=serve"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(2000 * time.Millisecond) // 等待 go func 启动完成
	a.NotNil(cmd.app).
		NotNil(cmd.app.getServer())

	p, err := os.FindProcess(os.Getpid())
	a.NotError(err).NotNil(p)

	// hup1
	t1 := cmd.app.getServer().Uptime()
	a.NotError(p.Signal(syscall.SIGHUP)).Wait(500 * time.Millisecond) // 此值要大于 CLI.ShutdownTimeout
	t2 := cmd.app.getServer().Uptime()
	a.True(t2.After(t1))

	// hup2
	a.NotError(p.Signal(syscall.SIGHUP)).Wait(500 * time.Millisecond) // 此值要大于 CLI.ShutdownTimeout
	t3 := cmd.app.getServer().Uptime()
	a.True(t3.After(t2))

	cmd.app.getServer().Close(0)
	<-exit
}
