// SPDX-License-Identifier: MIT

package dep

import (
	"errors"
	"log"
	"testing"

	"github.com/issue9/assert"
)

func TestModule_Tag(t *testing.T) {
	a := assert.New(t)

	m := &Module{tags: make(map[string]*Tag, 10)}
	t1 := m.Tag("t1")
	a.NotNil(t1)

	t2 := m.Tag("t2")
	a.NotNil(t2)

	t11 := m.Tag("t1")
	a.Equal(t11, t1)

	a.Equal(m.Tags(), []string{"t1", "t2"})
}

func TestTag_AddInit(t *testing.T) {
	a := assert.New(t)

	m := &Module{tags: make(map[string]*Tag, 10)}

	tag := m.Tag("t1")
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return nil }).
		AddInit("2", func() error { return nil })
	a.Equal(3, len(tag.executors))
	a.Equal(tag.executors[0].title, "1")
	a.Equal(tag.executors[2].title, "2")
}

func TestTag_init(t *testing.T) {
	a := assert.New(t)

	tag := &Tag{executors: make([]executor, 0, 5)}
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return nil })

	a.False(tag.Inited())
	a.NotError(tag.init(log.Default()))
	a.True(tag.Inited())
	a.NotError(tag.Inited())

	tag.AddInit("3", func() error { return nil })
	a.NotError(tag.init(log.Default()))
	a.True(tag.Inited())
	a.Equal(3, len(tag.executors))

	// failed

	tag = &Tag{executors: make([]executor, 0, 5)}
	tag.AddInit("1", func() error { return nil }).
		AddInit("2", func() error { return errors.New("error at 2") })
	a.False(tag.Inited())
	a.ErrorString(tag.init(log.Default()), "error at 2")
}
