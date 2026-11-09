// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v6"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

func TestSprint(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	a.Equal("", sprint(p, nil)).
		Equal("简体", sprint(p, web.Phrase("lang")))
}

func TestWriteMap2OrderedMap(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	ms := map[string]web.LocaleStringer{
		"t1": web.Phrase("lang"),
		"t2": web.Phrase("t2"),
	}

	om := writeMap2OrderedMap(ms, nil, func(in web.LocaleStringer) string { return sprint(p, in) })
	v1, found := om.Get("t1")
	v2, found := om.Get("t2")
	a.Equal(om.Len(), 2).
		True(found).Equal(v1, "简体").
		True(found).Equal(v2, "t2")

	om = writeMap2OrderedMap(nil, nil, func(in web.LocaleStringer) string { return sprint(p, in) })
	a.Nil(om)
}

func TestGetPathParams(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(getPathParams("/path/{id}/{id2}"), []string{"id", "id2"})
	a.Equal(getPathParams("/path/{id:number}/{id2}"), []string{"id:number", "id2"})
}
