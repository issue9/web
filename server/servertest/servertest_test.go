// SPDX-License-Identifier: MIT

package servertest

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestTester_Router(t *testing.T) {
	a := assert.New(t, false)

	s := NewServer(a, nil)
	r1 := s.Router()
	r2 := s.Router()
	a.Equal(r1, r2)
	defer s.Close(0)
}

func TestTester_Close(t *testing.T) {
	a := assert.New(t, false)

	s := NewServer(a, nil)
	s.Close(0)

	s = NewServer(a, nil)
	s.GoServe()
	s.Close(0)
}
