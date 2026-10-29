// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package startuptest

import (
	"testing"

	"github.com/issue9/assert/v6"
)

var _ Startup = &httpStartup{}

func TestHTTPStartup(t *testing.T) {
	t.Run("sec=false", func(t *testing.T) {
		a := assert.New(t, false)
		ts := NewHTTPServerStartup(a, false)
		s := newServer(a, ts)

		defer Run(a, s)()
		defer s.Close(0)

		testRequest(s, ts)
	})

	t.Run("sec=true", func(t *testing.T) {
		a := assert.New(t, false)
		ts := NewHTTPServerStartup(a, true)
		s := newServer(a, ts)

		defer Run(a, s)()
		defer s.Close(0)

		testRequest(s, ts)
	})
}
