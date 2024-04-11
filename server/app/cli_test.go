// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func TestCLI(t *testing.T) {
	a := assert.New(t, false)
	const shutdownTimeout = 0

	bs := new(bytes.Buffer)
	var action string
	cmd := &CLI[empty]{
		Name:            "test",
		Version:         "1.0.0",
		ConfigDir:       ".",
		ConfigFilename:  "web.yaml",
		ShutdownTimeout: shutdownTimeout,
		Out:             bs,
		ServeActions:    []string{"serve"},
		NewServer: func(name, ver string, opt *server.Options, _ *empty, act string) (web.Server, error) {
			action = act
			return server.New(name, ver, opt)
		},
	}
	a.NotError(cmd.Exec([]string{"app", "-v"})).Contains(bs.String(), cmd.Version)

	bs.Reset()
	a.NotError(cmd.Exec([]string{"app", "-a=install"})).Equal(action, "install")

	// RestartServer

	exit := make(chan struct{}, 10)
	bs.Reset()
	go func() {
		a.ErrorIs(cmd.Exec([]string{"app", "-a=serve"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(500 * time.Millisecond) // 等待 go func 启动完成

	// restart1
	s1 := cmd.app.getServer()
	t1 := s1.Uptime()
	cmd.Name = "restart1"
	cmd.RestartServer()
	time.Sleep(shutdownTimeout + 500*time.Millisecond) // 此值要大于 CLI.ShutdownTimeout
	s2 := cmd.app.getServer()
	t2 := s2.Uptime()
	a.True(t2.After(t1)).NotEqual(s1, s2)

	// restart2
	cmd.Name = "restart2"
	cmd.RestartServer()
	time.Sleep(shutdownTimeout + 500*time.Millisecond) // 此值要大于 CLI.ShutdownTimeout
	t3 := cmd.app.getServer().Uptime()
	a.True(t3.After(t2))

	cmd.app.getServer().Close(0)
	<-exit
}

func TestCLI_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &CLI[empty]{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &CLI[empty]{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "NewServer")

	cmd = &CLI[empty]{
		Name:    "app",
		Version: "1.1.1",
		NewServer: func(name, ver string, opt *server.Options, _ *empty, _ string) (web.Server, error) {
			return server.New(name, ver, opt)
		},
		ConfigFilename: "web.yaml",
	}
	a.NotError(cmd.sanitize()).Equal(cmd.Out, os.Stdout)

	cmd = &CLI[empty]{Name: "abc"}
	a.PanicString(func() {
		_ = cmd.Exec(nil)
	}, "字段 Version 不能为空")
}
