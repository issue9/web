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

func TestServer_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &Server{
		URL: "https://example.com",
	}
	a.NotError(s.valid())

	s = &Server{
		URL: "https://example.com/{id1}/{id2}",
		Variables: []*ServerVariable{
			{Name: "id1", Default: "1"},
			{Name: "id2", Default: "2"},
		},
	}
	a.NotError(s.valid())

	s = &Server{
		URL: "https://example.com/{id1}/{id2}",
		Variables: []*ServerVariable{
			{Name: "id1", Default: "1"},
		},
	}
	a.Equal(s.valid().Field, "Variables")
}

func TestServer_build(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	s := &Server{
		URL:         "https://example.com",
		Description: web.Phrase("lang"),
	}
	ret := s.build(p)
	a.Equal(ret.Description, "简体").Equal(ret.URL, s.URL)

	s = &Server{
		URL:         "https://example.com/{id1}/{id2}",
		Description: web.Phrase("lang"),
		Variables: []*ServerVariable{
			{Name: "id1", Default: "1", Description: web.Phrase("lang")},
			{Name: "id2", Default: "2", Description: web.Phrase("id2")},
		},
	}
	ret = s.build(p)
	a.Equal(ret.Description, "简体").Equal(ret.URL, s.URL).
		Equal(ret.Variables.Len(), 2).
		Equal(ret.Variables.GetPair("id1").Value.Description, "简体").
		Equal(ret.Variables.GetPair("id2").Value.Description, "id2")
}

func TestPathItem_addComponents(t *testing.T) {
	a := assert.New(t, false)
	d := New("1.0", web.Phrase("title"))

	item := &PathItem{
		Paths:   []*Parameter{{Name: "p1", Ref: &Ref{Ref: "p1"}}, {Name: "p2"}},
		Queries: []*Parameter{{Name: "q1", Ref: &Ref{Ref: "q1"}}, {Name: "q2"}},
		Headers: []*Parameter{{Name: "h1", Ref: &Ref{Ref: "h1"}}, {Name: "h2"}},
		Cookies: []*Parameter{{Name: "c1", Ref: &Ref{Ref: "c1"}}, {Name: "c2"}},
	}
	item.addComponents(d.components)
	a.Length(d.components.paths, 1).
		Length(d.components.cookies, 1).
		Length(d.components.queries, 1).
		Length(d.components.headers, 1)
}
