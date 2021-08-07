// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestServer_Publisher(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)

	p := s.Publisher("p1")
	a.NotNil(p)

	a.Panic(func() {
		s.Publisher("p1")
	})

	s.Publisher("")
	a.Panic(func() {
		s.Publisher("")
	})
}

func TestServer_Eventer(t *testing.T) {
	a := assert.New(t)
	s := newServer(a)
	buf := new(bytes.Buffer)

	p := s.Publisher("p1")
	a.NotNil(p)

	id1 := s.AttachEvent("p1", func(data interface{}) { fmt.Fprint(buf, data) })
	p.Publish("p1")
	time.Sleep(500 * time.Microsecond)
	a.Equal(buf.String(), "p1")

	s.DetachEvent("p1", id1)
	p.Publish("p1")
	time.Sleep(500 * time.Microsecond)
	a.Equal(buf.String(), "p1")
}
