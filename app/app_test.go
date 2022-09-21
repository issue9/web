// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"net/http"
	"os"
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
		Name:    "test",
		Version: "1.0.0",
		Out:     bs,
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
		a.ErrorIs(cmd.Exec([]string{"app", "-f=./testdata", "-s"}), http.ErrServerClosed)
		exit <- struct{}{}
	}()
	time.Sleep(500 * time.Millisecond) // 等待 go func 启动完成

	s1 := cmd.srv
	t1 := s1.Uptime()
	cmd.Name = "restart"
	cmd.Restart()
	time.Sleep(500 * time.Microsecond) //////////// TODO
	s2 := cmd.srv
	t2 := s2.Uptime()
	a.True(t2.After(t1)).
		NotEqual(s1, s2)

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
}
