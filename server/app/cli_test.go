// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/issue9/assert/v5"

	"github.com/issue9/cmdopt"
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
		NewServer: func(name, ver string, opt *server.Options, _ empty) (web.Server, error) {
			return server.NewHTTP(name, ver, opt)
		},
		Commands: []*cmdopt.Command{
			{Name: "install", Title: "title", Usage: "usage", Command: func(*flag.FlagSet) cmdopt.DoFunc {
				return func(io.Writer) error {
					action = "install"
					return nil
				}
			}},
		},
	}
	cmd := NewCLI(o)
	ocli := cmd.(*cli[empty])
	a.NotError(ocli.exec([]string{"app", "-v"})).Contains(buf.String(), o.Version)

	buf.Reset()
	a.NotError(ocli.exec([]string{"app", "install"})).Equal(action, "install")

	buf.Reset()
	msg := web.Phrase("syntax OK").LocaleString(o.Printer) + "\n"
	a.NotError(ocli.exec([]string{"app", "-t"})).Equal(buf.String(), msg)
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
		NewServer: func(name, ver string, opt *server.Options, _ empty) (web.Server, error) {
			return server.NewHTTP(name, ver, opt)
		},
		ConfigFilename: "web.yaml",
	}
	a.NotError(cmd.sanitize()).Equal(cmd.Out, os.Stdout)

	a.PanicString(func() {
		NewCLI(&CLIOptions[empty]{})
	}, "ID")
}
