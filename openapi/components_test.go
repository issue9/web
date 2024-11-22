// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

func TestParameter_addComponents(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	p := &Parameter{}
	p.addComponents(d.components, InPath)
	a.Empty(d.paths)

	p = &Parameter{Schema: &Schema{Type: TypeString}}
	p.addComponents(d.components, InPath)
	a.Empty(d.paths)

	p = &Parameter{Schema: &Schema{Type: TypeString}, Ref: &Ref{Ref: "string"}}
	p.addComponents(d.components, InPath)
	a.Equal(d.components.paths["string"], p)
}

func TestSchema_addComponents(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	s := &Schema{}
	s.addComponents(d.components)
	a.Empty(d.components.schemas)

	s = &Schema{Type: TypeString}
	s.addComponents(d.components)
	a.Empty(d.components.schemas)

	s1 := &Schema{Type: TypeString, Ref: &Ref{Ref: "t1"}}
	s1.addComponents(d.components)
	a.Length(d.components.schemas, 1).Equal(d.components.schemas["t1"], s1)

	// 同名不会再添加
	s2 := &Schema{Type: TypeString, Ref: &Ref{Ref: "t1"}}
	s2.addComponents(d.components)
	a.Length(d.components.schemas, 1).Equal(d.components.schemas["t1"], s1)

	s2 = &Schema{Type: TypeString, Ref: &Ref{Ref: "t2"}, Items: s1}
	s2.addComponents(d.components)
	a.Length(d.components.schemas, 2).Equal(d.components.schemas["t2"], s2)

	s3 := &Schema{Type: TypeString, Ref: &Ref{Ref: "t3"}, Items: &Schema{Type: TypeNumber, Ref: &Ref{Ref: "t4"}}}
	s3.addComponents(d.components)
	a.Length(d.components.schemas, 4).
		Equal(d.components.schemas["t3"], s3).
		Equal(d.components.schemas["t4"].Type, TypeNumber)
}

func TestPathItem_addComponents(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"))

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
