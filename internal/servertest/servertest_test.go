// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package servertest

import (
	"strings"
	"testing"

	"github.com/issue9/assert/v6"
)

func TestNewServer(t *testing.T) {
	t.Run("sec=false", func(t *testing.T) {
		a := assert.New(t, false)
		s := NewServer(a, false)
		a.Equal(s.A, a).
			Empty(s.Certificates).
			True(strings.HasPrefix(s.URL, "http://"))
	})

	t.Run("sec=true", func(t *testing.T) {
		a := assert.New(t, false)
		s := NewServer(a, true)
		a.Equal(s.A, a).
			NotEmpty(s.Certificates).
			True(strings.HasPrefix(s.URL, "https://"))
	})
}
