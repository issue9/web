// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

func TestCLI(t *testing.T) {
	a := assert.New(t, false)
	const shutdownTimeout = 0

	bs := new(bytes.Buffer)
	var action string
	o := &CLIOptions[empty]{
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
	cmd := NewCLI(o)
	ocli := cmd.(*cli[empty])
	a.NotError(ocli.exec([]string{"app", "-v"})).Contains(bs.String(), o.Version)

	bs.Reset()
	a.NotError(ocli.exec([]string{"app", "-a=install"})).Equal(action, "install")
}

func TestCLI_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &CLIOptions[empty]{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &CLIOptions[empty]{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "NewServer")

	cmd = &CLIOptions[empty]{
		Name:    "app",
		Version: "1.1.1",
		NewServer: func(name, ver string, opt *server.Options, _ *empty, _ string) (web.Server, error) {
			return server.New(name, ver, opt)
		},
		ConfigFilename: "web.yaml",
	}
	a.NotError(cmd.sanitize()).Equal(cmd.Out, os.Stdout)

	a.PanicValue(func() {
		NewCLI(&CLIOptions[empty]{Name: "abc"})
	}, web.NewFieldError("Version", locales.ErrCanNotBeEmpty()))
}
