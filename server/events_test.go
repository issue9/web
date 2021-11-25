// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
)

func TestServer_Publisher(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)

	p := s.Publisher("p1")
	a.NotNil(p)

	a.Panic(func() {
		s.Publisher("p1")
	})

	s.Publisher("")
	a.Panic(func() {
		s.Publisher("")
	})

	p.Destory()
	a.Nil(s.events["p1"])
}

func TestServer_Eventer(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	buf := new(bytes.Buffer)

	p := s.Publisher("p1")
	a.NotNil(p)

	id1, err := s.AttachEvent("p1", func(data interface{}) { fmt.Fprint(buf, data) })
	a.NotError(err)
	a.NotError(p.Publish(true, "p1"))
	time.Sleep(500 * time.Microsecond)
	a.Equal(buf.String(), "p1")

	s.DetachEvent("p1", id1)
	a.NotError(p.Publish(false, "p1"))
	time.Sleep(500 * time.Microsecond)
	a.Equal(buf.String(), "p1")
}
