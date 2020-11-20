// SPDX-License-Identifier: MIT

package dep

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/issue9/assert"
)

var _ Module = &Default{}

func newMod(id string, f func() error, dep ...string) *Default {
	d := NewDefaultModule(id, id+" description", dep...)
	d.AddInit(f, id)
	return d
}

func TestDefault_AddInit(t *testing.T) {
	a := assert.New(t)

	m := NewDefaultModule("m1", "m1 dexc")

	a.Empty(m.inits)
	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 1).
		Equal(m.inits[0].title, "t1").
		NotNil(m.inits[0].f)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 2).
		Equal(m.inits[1].title, "t1").
		NotNil(m.inits[1].f)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 3).
		Equal(m.inits[2].title, "t1").
		NotNil(m.inits[2].f)

	a.NotError(m.Init(log.New(ioutil.Discard, "", 0)))
	a.Panic(func() {
		m.AddInit(func() error { return nil }, "t1")
	})
}
