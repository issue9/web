// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"github.com/issue9/assert"
)

func TestIn(t *testing.T) {
	a := assert.New(t)

	rule := In("msg", 1, 2, "3", struct{}{})
	a.Equal(rule.Validate(3), "msg")
	a.Equal(rule.Validate("1"), "msg")
	a.Empty(rule.Validate(1))
	a.Empty(rule.Validate(uint8(1)))

	// object
	type obj struct {
		Name string
		Age  int
	}
	rule = In("msg", 1, "2", &obj{}, &obj{Name: "name"})
	a.Equal(rule.Validate(3), "msg")
	a.Equal(rule.Validate("1"), "msg")
	a.Empty(rule.Validate(&obj{}))
	a.Empty(rule.Validate(&obj{Name: "name"}))
	a.Equal(rule.Validate(&obj{Name: "name", Age: 1}), "msg")
}

func TestNotIn(t *testing.T) {
	a := assert.New(t)

	rule := NotIn("msg", 1, 2, "3", struct{}{})
	a.Empty(rule.Validate(3))
	a.Empty(rule.Validate("1"))
	a.Equal(rule.Validate(1), "msg")
	a.Equal(rule.Validate(uint8(1)), "msg")

	// object
	type obj struct {
		Name string
		Age  int
	}
	rule = NotIn("msg", 1, "2", &obj{}, &obj{Name: "name"})
	a.Empty(rule.Validate(3))
	a.Empty(rule.Validate("1"))
	a.Equal(rule.Validate(&obj{}), "msg")
	a.Equal(rule.Validate(&obj{Name: "name"}), "msg")
	a.Empty(rule.Validate(&obj{Name: "name", Age: 1}))
}
