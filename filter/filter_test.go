// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	v := "str "
	f := New("name", &v, trimRight, NewRuleOf(Not(zero[string]), localeutil.Phrase("zero")))
	name, msg := f()
	a.Empty(name).
		Nil(msg).
		Equal(v, "str")

	v = ""
	name, msg = f()
	a.Equal(name, "name").
		Equal(msg, localeutil.Phrase("zero"))
}
