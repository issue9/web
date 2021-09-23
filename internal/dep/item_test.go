// SPDX-License-Identifier: MIT

package dep

import (
	"testing"

	"github.com/issue9/assert"
)

func TestReverse(t *testing.T) {
	a := assert.New(t)

	m1 := &Item{
		ID:        "m1",
		Deps:      []string{"m2", "m3"},
		Executors: []Executor{{Title: "init m1", F: func() error { return nil }}},
	}
	m2 := &Item{
		ID:        "m2",
		Deps:      []string{"m3"},
		Executors: []Executor{{Title: "init m2", F: func() error { return nil }}},
	}
	m3 := &Item{
		ID:        "m3",
		Executors: []Executor{{Title: "init m3", F: func() error { return nil }}},
	}
	items := []*Item{m1, m2, m3}

	ret := Reverse(items)
	m := findItem(ret, "m1")
	a.NotError(m).Equal(m.Deps, []string{})
	a.NotEmpty(m1.Executors)

	m = findItem(ret, "m2")
	a.NotError(m).Equal(m.Deps, []string{"m1"})

	m = findItem(ret, "m3")
	a.NotError(m).Equal(m.Deps, []string{"m1", "m2"})
}
