// SPDX-License-Identifier: MIT

package web

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestCommand_sanitize(t *testing.T) {
	a := assert.New(t)

	cmd := &Command{}
	err := cmd.sanitize()
	a.Equal(err.Field, "Name")

	cmd = &Command{Name: "app", Version: "1.1.1"}
	err = cmd.sanitize()
	a.Equal(err.Field, "InitServer")

	cmd = &Command{Name: "app", Version: "1.1.1", InitServer: func(s *Server) error { return nil }}
	err = cmd.sanitize()
	a.NotError(err)

	a.Equal(cmd.Out, os.Stdout).
		NotNil(cmd.Locale).
		Equal(cmd.CmdFS, "fs")
}
