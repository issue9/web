// SPDX-License-Identifier: MIT

package git

import (
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestVersion(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	v := Version(p)
	a.NotEmpty(v)

	f := Commit(p, true)
	a.NotEmpty(f)

	s := Commit(p, false)
	a.NotEmpty(s).True(strings.HasPrefix(f, s), "v1=%s,v2=%s", f, s)
}
