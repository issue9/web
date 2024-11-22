// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

func TestDocument_build(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)

	d := New(s, web.Phrase("lang"))
	r := d.build(p, language.SimplifiedChinese, nil)
	a.Equal(r.Info.Version, s.Version()).
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体")

	d.addOperation("GET", "/users/{id}", "", &Operation{
		Paths:     []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
		Responses: map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
	})
	r = d.build(p, language.SimplifiedChinese, nil)
	a.Equal(r.Info.Version, s.Version()).
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体").
		Equal(r.Paths.Len(), 1)

	d.addOperation("POST", "/users/{id}", "", &Operation{
		Tags:      []string{"admin"},
		Paths:     []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
		Responses: map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
	})
	r = d.build(p, language.SimplifiedChinese, nil)
	obj := r.Paths.GetPair("/users/{id}").Value.obj
	a.NotNil(obj.Get).
		NotNil(obj.Post).
		Nil(obj.Delete)

	// 带过滤

	r = d.build(p, language.SimplifiedChinese, []string{"admin"})
	obj = r.Paths.GetPair("/users/{id}").Value.obj
	a.Nil(obj.Get).
		NotNil(obj.Post).
		Nil(obj.Delete)
}

func TestComponents_build(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)
	d := New(s, web.Phrase("lang"))
	c := d.components

	c.queries["q1"] = &Parameter{Name: "q1", Schema: &Schema{Type: TypeString}}
	c.cookies["c1"] = &Parameter{Name: "c1", Schema: &Schema{Type: TypeNumber}}
	c.headers["h1"] = &Parameter{Name: "h1", Schema: &Schema{Type: TypeBoolean}}
	c.schemas["s1"] = NewSchema(reflect.TypeFor[int](), nil, nil)

	r := c.build(p, d)
	a.Equal(r.Parameters.Len(), 2).
		Equal(r.Headers.Len(), 1).
		Equal(r.Schemas.Len(), 1)
}

func TestServer_build(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

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

func TestRef_build(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	ref := &Ref{}
	a.PanicString(func() {
		ref.build(p, "schemas")
	}, "ref 不能为空")

	ref = &Ref{
		Ref:         "ref",
		Summary:     web.Phrase("lang"),
		Description: web.Phrase("desc"),
	}
	a.Equal(ref.build(p, "schemas"), &refRenderer{Ref: "#/components/schemas/ref", Summary: "简体", Description: "desc"})
}

func TestSchema_build(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	var s *Schema
	a.Nil(s.build(p))

	s = &Schema{Ref: &Ref{Ref: "ref"}, Type: TypeArray}
	sr := s.build(p)
	a.Equal(sr.ref.Ref, "#/components/schemas/ref").Nil(sr.obj)

	s = &Schema{Type: TypeArray, Description: web.Phrase("lang")}
	sr = s.build(p)
	a.Nil(sr.ref).NotNil(sr.obj).
		Equal(sr.obj.Type, TypeArray).
		Equal(sr.obj.Description, "简体")
}

func TestSecurityScheme_build(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	s := &SecurityScheme{
		Description: web.Phrase("lang"),
		Type:        SecuritySchemeTypeOAuth2,
		Flows: &OAuthFlows{
			Password: &OAuthFlow{
				TokenURL: "https://example.com/token",
			},
		},
	}

	r := s.build(p)
	a.NotNil(r).Equal(r.Description, "简体").
		NotNil(r.Flows).
		NotNil(r.Flows.Password).
		Nil(r.Flows.Implicit)
}
