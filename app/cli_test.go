// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
)

func TestCLIOf(t *testing.T) {
	a := assert.New(t, false)
	const shutdownTimeout = 0

	bs := new(bytes.Buffer)
	var action string
	cmd := &CLIOf[empty]{
		Name:            "test",
		Version:         "1.0.0",
		ConfigDir:       "./testdata",
		ConfigFilename:  "web.yaml",
		ShutdownTimeout: shutdownTimeout,
		Out:             bs,
		ServeActions:    []string{"serve"},
		Init: func(s *web.Server, user *empty, act string) error {
			action = act
			return nil
		},
	}
	a.NotError(cmd.Exec([]string{"app", "-v"}))
	a.Contains(bs.String(), cmd.Version)

	bs.Reset()
	a.NotError(cmd.Exec([]string{"app", "-a=install"}))
	a.Equal(action, "install")

	// RestartServer

	exit := make(chan struct{}, 10)
	bs.Reset()
	go func() {
		a.ErrorIs(cmd.Exec([]string{"app", "-a=serve"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(500 * time.Millisecond) // 等待 go func 启动完成

	// restart1
	s1 := cmd.app.srv
	t1 := s1.Uptime()
	cmd.Name = "restart1"
	cmd.RestartServer()
	time.Sleep(shutdownTimeout + 500*time.Millisecond) // 此值要大于 CLIOf.ShutdownTimeout
	s2 := cmd.app.srv
	t2 := s2.Uptime()
	a.True(t2.After(t1)).NotEqual(s1, s2)

	// restart2
	cmd.Name = "restart2"
	cmd.RestartServer()
	time.Sleep(shutdownTimeout + 500*time.Millisecond) // 此值要大于 CLIOf.ShutdownTimeout
	t3 := cmd.app.srv.Uptime()
	a.True(t3.After(t2))

	cmd.app.srv.Close(0)
	<-exit
}

func TestCLIOf_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &CLIOf[empty]{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &CLIOf[empty]{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &CLIOf[empty]{
		Name:           "app",
		Version:        "1.1.1",
		Init:           func(*web.Server, *empty, string) error { return nil },
		ConfigFilename: "web.yaml",
	}
	a.NotError(cmd.sanitize())
	a.Equal(cmd.Out, os.Stdout)

	cmd = &CLIOf[empty]{Name: "abc"}
	a.PanicString(func() {
		cmd.Exec(nil)
	}, "字段 Version 不能为空")
}
