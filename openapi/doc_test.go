// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

func TestParameterizedDoc(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)
	d := New(s, web.Phrase("title"))

	a.PanicString(func() {
		d.ParameterizedDoc("format", nil)
	}, "参数 format 必须包含 %s")

	ls := d.ParameterizedDoc("format %s", web.Phrase("p1\n"))
	a.Length(ls.(*parameterizedDoc).params, 1).
		Length(d.parameterizedDocs, 1)

	ls = d.ParameterizedDoc("format %s", web.Phrase("p2\n"))
	a.Length(ls.(*parameterizedDoc).params, 2).
		Equal(ls.LocaleString(p), web.Phrase("format %s", "p1\np2\n").LocaleString(p))
}
