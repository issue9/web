// SPDX-License-Identifier: MIT

package validation

import (
	"regexp"
	"testing"

	"github.com/issue9/assert"
)

func TestMatch(t *testing.T) {
	a := assert.New(t)

	r := Match("msg", regexp.MustCompile("[a-z]+"))
	a.Empty(r.Validate("abc"))
	a.Empty(r.Validate([]byte("def")))
	a.Equal(r.Validate([]rune("123")), "msg")
	a.Equal(r.Validate(123), "msg") // 无法验证
}
