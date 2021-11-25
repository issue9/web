// SPDX-License-Identifier: MIT

package web

import (
	"os"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestCommand_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &Command{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &Command{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &Command{Name: "app", Version: "1.1.1", Init: func(*Server, bool, string) error { return nil }}
	a.NotError(cmd.sanitize())

	a.Equal(cmd.Out, os.Stdout)
}
