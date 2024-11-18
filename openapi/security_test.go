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
