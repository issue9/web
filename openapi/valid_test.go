// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestServer_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &Server{
		URL: "https://example.com",
	}
	m, f := createFilter()
	s.valid(f)
	a.False(f.HasError()).Empty(m)

	s = &Server{
		URL: "https://example.com/{id1}/{id2}",
		Variables: []*ServerVariable{
			{Name: "id1", Default: "1"},
			{Name: "id2", Default: "2"},
		},
	}
	m, f = createFilter()
	s.valid(f)
	a.False(f.HasError()).Empty(m)

	s = &Server{
		URL: "https://example.com/{id1}/{id2}",
		Variables: []*ServerVariable{
			{Name: "id1", Default: "1"},
		},
	}
	m, f = createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["Variables"])
}

func TestParameter_valid(t *testing.T) {
	a := assert.New(t, false)

	p := &Parameter{}
	m, f := createFilter()
	p.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Name"])

	p = &Parameter{Name: "p1"}
	m, f = createFilter()
	p.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Schema"])

	p = &Parameter{Name: "p1", Schema: &Schema{}}
	m, f = createFilter()
	p.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Schema.Type"])

	p = &Parameter{Name: "p1", Schema: &Schema{Type: TypeString}}
	m, f = createFilter()
	p.valid(f, false)
	a.False(f.HasError()).Empty(m)

	p = &Parameter{Name: "p1", Schema: &Schema{Type: TypeObject}}
	m, f = createFilter()
	p.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Schema"])
}

func TestSchema_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &Schema{}
	m, f := createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Type"])

	s = &Schema{Type: TypeString}
	m, f = createFilter()
	s.valid(f, false)
	a.False(f.HasError()).Empty(m)

	s = &Schema{Type: TypeString, Properties: map[string]*Schema{"f1": {Type: TypeString}}}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Type"])

	s = &Schema{Type: TypeObject, Properties: map[string]*Schema{"f1": {}}}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Properties[f1].Type"])

	s = &Schema{Type: TypeObject, Properties: map[string]*Schema{}} // 空对象是合法的
	m, f = createFilter()
	s.valid(f, false)
	a.False(f.HasError()).Empty(m)

	s = &Schema{Type: TypeString, Items: &Schema{Type: TypeString}}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Type"])

	s = &Schema{Type: TypeArray}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["Type"])

	s = &Schema{Type: TypeString, AllOf: []*Schema{{}}}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["AllOf[0].Type"])

	s = &Schema{Type: TypeString, OneOf: []*Schema{{}}}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["OneOf[0].Type"])

	s = &Schema{Type: TypeString, AnyOf: []*Schema{{Type: TypeString}, {}}}
	m, f = createFilter()
	s.valid(f, false)
	a.True(f.HasError()).NotEmpty(m["AnyOf[1].Type"])
}

func TestSecurityScheme_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &SecurityScheme{}
	m, f := createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["Type"])

	s = &SecurityScheme{Type: SecuritySchemeTypeOAuth2}
	m, f = createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["Flows"])

	s = &SecurityScheme{Type: SecuritySchemeTypeHTTP}
	m, f = createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["Scheme"])

	s = &SecurityScheme{Type: SecuritySchemeTypeOpenIDConnect}
	m, f = createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["OpenIDConnectURL"])

	s = &SecurityScheme{Type: SecuritySchemeTypeAPIKey}
	m, f = createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["Name"])

	s = &SecurityScheme{Type: SecuritySchemeTypeAPIKey, Name: "token"}
	m, f = createFilter()
	s.valid(f)
	a.True(f.HasError()).NotEmpty(m["In"])
}
