// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package startuptest

import (
	"testing"

	"github.com/issue9/assert/v6"
)

var _ Startup = &testStartup{}

func TestTestStartup(t *testing.T) {
	t.Run("NewTestServerStartup", func(t *testing.T) {
		a := assert.New(t, false)
		testTestStartup(a, NewTestServerStartup(a))
	})

	t.Run("NewServerStartup", func(t *testing.T) {
		a := assert.New(t, false)
		testTestStartup(a, NewServerStartup(a))
	})

	t.Run("NewTLSServerStartup", func(t *testing.T) {
		a := assert.New(t, false)
		testTestStartup(a, NewTLSServerStartup(a))
	})
}

func testTestStartup(a *assert.Assertion, ts Startup) {
	s := newServer(a, ts)
	testRequest(s, ts)
}
