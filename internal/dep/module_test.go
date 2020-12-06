// SPDX-License-Identifier: MIT

package dep

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/issue9/assert"
)

func newMod(id string, f func() error, dep ...string) *Module {
	d := NewModule(id, id+" description", dep...)
	d.AddInit(id, f)
	return d
}

func TestDefault_AddInit(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m1", "m1 dexc")

	a.Empty(m.inits)
	m.AddInit("t1", func() error { return nil })
	a.Equal(len(m.inits), 1).
		Equal(m.inits[0].title, "t1").
		NotNil(m.inits[0].f)

	m.AddInit("t1", func() error { return nil })
	a.Equal(len(m.inits), 2).
		Equal(m.inits[1].title, "t1").
		NotNil(m.inits[1].f)

	m.AddInit("t1", func() error { return nil })
	a.Equal(len(m.inits), 3).
		Equal(m.inits[2].title, "t1").
		NotNil(m.inits[2].f)

	a.NotError(m.Init(log.New(ioutil.Discard, "", 0)))
	a.Panic(func() {
		m.AddInit("t1", func() error { return nil })
	})
}
