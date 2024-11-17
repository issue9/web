// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

func TestSprint(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	a.Equal("", sprint(p, nil)).
		Equal("简体", sprint(p, web.Phrase("lang")))
}

func TestWriteMap2OrderedMap(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	ms := map[string]web.LocaleStringer{
		"t1": web.Phrase("lang"),
		"t2": web.Phrase("t2"),
	}

	om := writeMap2OrderedMap(ms, nil, func(in web.LocaleStringer) string { return sprint(p, in) })
	a.Equal(om.Len(), 2).
		Equal(om.GetPair("t1").Value, "简体").
		Equal(om.GetPair("t2").Value, "t2")

	om = writeMap2OrderedMap[string, web.LocaleStringer, string](nil, nil, func(in web.LocaleStringer) string { return sprint(p, in) })
	a.Nil(om)
}

func TestGetPathParams(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(getPathParams("/path/{id}/{id2}"), []string{"id", "id2"})
	a.Equal(getPathParams("/path/{id:number}/{id2}"), []string{"id:number", "id2"})
}
