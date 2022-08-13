// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/server"
)

func TestAppOf_Exec(t *testing.T) {
	a := assert.New(t, false)

	bs := new(bytes.Buffer)
	var action string
	aa := &AppOf[empty]{
		Name:    "test",
		Version: "1.0.0",
		Out:     bs,
		Init: func(s *server.Server, user *empty, act string) error {
			action = act
			return nil
		},
	}
	a.NotError(aa.Exec([]string{"app", "-v"}))
	a.Contains(bs.String(), aa.Version)

	bs.Reset()
	a.NotError(aa.Exec([]string{"app", "-f=./testdata", "-a=install"}))
	a.Equal(action, "install")
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
