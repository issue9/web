// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"
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

func TestParameter_valid(t *testing.T) {
	a := assert.New(t, false)

	p := &Parameter{}
	err := p.valid(false)
	a.Equal(err.Field, "Name")

	p = &Parameter{Name: "p1"}
	err = p.valid(false)
	a.Equal(err.Field, "Schema")

	p = &Parameter{Name: "p1", Schema: &Schema{}}
	err = p.valid(false)
	a.Equal(err.Field, "Schema.Type")

	p = &Parameter{Name: "p1", Schema: &Schema{Type: TypeString}}
	a.NotError(p.valid(false))

	p = &Parameter{Name: "p1", Schema: &Schema{Type: TypeObject}}
	err = p.valid(false)
	a.Equal(err.Field, "Schema")
}

func TestSchema_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &Schema{}
	err := s.valid(false)
	a.Equal(err.Field, "Type")

	s = &Schema{Type: TypeString}
	a.NotError(s.valid(false))

	s = &Schema{Type: TypeString, Properties: map[string]*Schema{"f1": {Type: TypeString}}}
	err = s.valid(false)
	a.Equal(err.Field, "Type")
	s = &Schema{Type: TypeObject, Properties: map[string]*Schema{"f1": {}}}
	err = s.valid(false)
	a.Equal(err.Field, "Properties[f1].Type")
	s = &Schema{Type: TypeObject, Properties: map[string]*Schema{}} // 空对象是合法的
	a.Nil(s.valid(false))

	s = &Schema{Type: TypeString, Items: &Schema{Type: TypeString}}
	err = s.valid(false)
	a.Equal(err.Field, "Type")
	s = &Schema{Type: TypeArray}
	err = s.valid(false)
	a.Equal(err.Field, "Type")

	s = &Schema{Type: TypeString, AllOf: []*Schema{{}}}
	err = s.valid(false)
	a.Equal(err.Field, "AllOf[0].Type")

	s = &Schema{Type: TypeString, OneOf: []*Schema{{}}}
	err = s.valid(false)
	a.Equal(err.Field, "OneOf[0].Type")

	s = &Schema{Type: TypeString, AnyOf: []*Schema{{Type: TypeString}, {}}}
	err = s.valid(false)
	a.Equal(err.Field, "AnyOf[1].Type")
}

func TestSecurityScheme_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &SecurityScheme{}
	a.Equal(s.valid().Field, "Type")

	s = &SecurityScheme{Type: SecuritySchemeTypeOAuth2}
	a.Equal(s.valid().Field, "Flows")

	s = &SecurityScheme{Type: SecuritySchemeTypeHTTP}
	a.Equal(s.valid().Field, "Scheme")

	s = &SecurityScheme{Type: SecuritySchemeTypeOpenIDConnect}
	a.Equal(s.valid().Field, "OpenIDConnectURL")

	s = &SecurityScheme{Type: SecuritySchemeTypeAPIKey}
	a.Equal(s.valid().Field, "Name")
	s = &SecurityScheme{Type: SecuritySchemeTypeAPIKey, Name: "token"}
	a.Equal(s.valid().Field, "In")
}
