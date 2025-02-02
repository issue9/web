// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

var _ App = &cli[int]{}

func TestCLI(t *testing.T) {
	a := assert.New(t, false)
	const shutdownTimeout = 0

	buf := new(bytes.Buffer)
	var action string
	o := &CLIOptions[empty]{
		ID:              "test",
		Version:         "1.0.0",
		ConfigDir:       ".",
		ConfigFilename:  "web.yaml",
		ShutdownTimeout: shutdownTimeout,
		Out:             buf,
		ServeActions:    []string{"serve"},
		NewServer: func(name, ver string, opt *server.Options, _ empty, act string) (web.Server, error) {
			action = act
			return server.NewHTTP(name, ver, opt)
		},
	}
	cmd := NewCLI(o)
	ocli := cmd.(*cli[empty])
	a.NotError(ocli.exec([]string{"app", "-v"})).Contains(buf.String(), o.Version)

	buf.Reset()
	a.NotError(ocli.exec([]string{"app", "-a=install"})).Equal(action, "install")
}

func TestCLI_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &CLIOptions[empty]{}
	a.ErrorString(cmd.sanitize(), "ID")

	cmd = &CLIOptions[empty]{ID: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "NewServer")

	cmd = &CLIOptions[empty]{
		ID:      "app",
		Version: "1.1.1",
		NewServer: func(name, ver string, opt *server.Options, _ empty, _ string) (web.Server, error) {
			return server.NewHTTP(name, ver, opt)
		},
		ConfigFilename: "web.yaml",
	}
	a.NotError(cmd.sanitize()).Equal(cmd.Out, os.Stdout)

	a.PanicString(func() {
		NewCLI(&CLIOptions[empty]{ID: "abc"})
	}, "Version")
}
