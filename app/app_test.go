// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/server"
)

func TestAppOf(t *testing.T) {
	a := assert.New(t, false)

	bs := new(bytes.Buffer)
	var action string
	cmd := &AppOf[empty]{
		Name:           "test",
		Version:        "1.0.0",
		ConfigFilename: "web.yaml",
		Out:            bs,
		ServeActions:   []string{"serve"},
		Init: func(s *server.Server, user *empty, act string) error {
			action = act
			return nil
		},
	}
	a.NotError(cmd.Exec([]string{"app", "-v"}))
	a.Contains(bs.String(), cmd.Version)

	bs.Reset()
	a.NotError(cmd.Exec([]string{"app", "-f=./testdata", "-a=install"}))
	a.Equal(action, "install")

	// Restart

	exit := make(chan struct{}, 10)
	bs.Reset()
	go func() {
		a.ErrorIs(cmd.Exec([]string{"app", "-f=./testdata", "-a=serve"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(500 * time.Millisecond) // 等待 go func 启动完成

	// restart1
	s1 := cmd.srv
	t1 := s1.Uptime()
	cmd.Name = "restart1"
	cmd.Restart()
	time.Sleep(500 * time.Millisecond) // 此值要大于 AppOf.ShutdownTimeout
	s2 := cmd.srv
	t2 := s2.Uptime()
	a.True(t2.After(t1)).
		NotEqual(s1, s2)

	// restart2
	cmd.Name = "restart2"
	cmd.Restart()
	time.Sleep(500 * time.Millisecond) // 此值要大于 AppOf.ShutdownTimeout
	t3 := cmd.srv.Uptime()
	a.True(t3.After(t2))

	a.NotError(cmd.srv.Close(0))
	<-exit
}

func TestSIGHUP(t *testing.T) {
	a := assert.New(t, false)

	exit := make(chan struct{}, 10)
	cmd := &AppOf[empty]{
		Name:           "test",
		Version:        "1.0.0",
		ConfigFilename: "web.yaml",
		ServeActions:   []string{"serve"},
		Init: func(s *server.Server, user *empty, act string) error {
			return nil
		},
	}

	go func() {
		a.ErrorIs(cmd.Exec([]string{"app", "-f=./testdata", "-a=serve"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(500 * time.Millisecond) // 等待 go func 启动完成

	if runtime.GOOS != "windows" {
		SignalHUP(cmd)

		p, err := os.FindProcess(os.Getpid())
		a.NotError(err).NotNil(p)

		// hup1
		t1 := cmd.srv.Uptime()
		cmd.Name = "hup1"
		a.NotError(p.Signal(syscall.SIGHUP))
		time.Sleep(500 * time.Millisecond) // 此值要大于 AppOf.ShutdownTimeout
		t2 := cmd.srv.Uptime()
		a.True(t2.After(t1)).Equal(cmd.srv.Name(), "hup1")

		// hup2
		cmd.Name = "hup2"
		a.NotError(p.Signal(syscall.SIGHUP))
		time.Sleep(500 * time.Millisecond) // 此值要大于 AppOf.ShutdownTimeout
		t3 := cmd.srv.Uptime()
		a.True(t3.After(t2)).Equal(cmd.srv.Name(), "hup2")
	}

	a.NotError(cmd.srv.Close(0))
	<-exit
}

func TestAppOf_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &AppOf[empty]{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &AppOf[empty]{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &AppOf[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
	}
	a.NotError(cmd.sanitize())
	a.Equal(cmd.Out, os.Stdout)

	cmd = &AppOf[empty]{Name: "abc"}
	a.PanicString(func() {
		cmd.Exec(nil)
	}, "字段 Version 不能为空")
}
