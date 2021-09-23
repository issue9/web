// SPDX-License-Identifier: MIT

package dep

import (
	"testing"

	"github.com/issue9/assert"
)

func TestReverse(t *testing.T) {
	a := assert.New(t)

	m1 := NewItem("m1", []string{"m2", "m3"}, []Executor{{Title: "init m1", F: func() error { return nil }}})
	m2 := NewItem("m2", []string{"m3"}, []Executor{{Title: "init m2", F: func() error { return nil }}})
	m3 := NewItem("m3", nil, []Executor{{Title: "init m3", F: func() error { return nil }}})
	items := []*Item{m1, m2, m3}

	ret := Reverse(items)
	m := findItem(ret, "m1")
	a.NotError(m).Equal(m.deps, []string{})
	a.NotEmpty(m1.executors)

	m = findItem(ret, "m2")
	a.NotError(m).Equal(m.deps, []string{"m1"})

	m = findItem(ret, "m3")
	a.NotError(m).Equal(m.deps, []string{"m1", "m2"})
}
