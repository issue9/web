// SPDX-License-Identifier: MIT

package module

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v3"
)

var _ Initializer = &initializer{}

func buildFailedInit(msg string) func() error {
	return func() error {
		return errors.New(msg)
	}
}

func TestInitializer_AddInit(t *testing.T) {
	a := assert.New(t)

	i := &initializer{name: "name-1"}

	a.PanicString(func() {
		i.AddInit("", func() error { return nil })
	}, "name")
	a.Empty(i.inits)

	a.PanicString(func() {
		i.AddInit("n1", nil)
	}, "f")
	a.Empty(i.inits)

	a.NotError(i.AddInit("n1", func() error { return nil }))
	a.Equal(1, len(i.inits))
	a.Equal("n1", i.inits[0].name)

	// 可以同名
	a.NotError(i.AddInit("n1", func() error { return nil }))
	a.Equal(2, len(i.inits))
	a.Equal("n1", i.inits[1].name)
}

func TestInitializer_init(t *testing.T) {
	a := assert.New(t)
	l, err := logs.New(nil)
	a.NotError(err).NotNil(l)

	i := &initializer{name: "name-1"}
	a.NotError(i.init(l, 0))

	a.NotNil(i.AddInit("f1", buildFailedInit("f1-failed")))
	a.ErrorString(i.init(l, 0), "f1-failed")

	i = &initializer{name: "name-1"}
	ii := i.AddInit("f2", func() error { return nil })
	a.NotNil(ii)
	a.NotNil(ii.AddInit("f3", buildFailedInit("f3-failed")))
	a.ErrorString(i.init(l, 0), "f3-failed")
}
