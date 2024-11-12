// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

type (
	SecurityScheme struct {
		Type             string
		Description      web.LocaleStringer
		Name             web.LocaleStringer
		In               string
		Scheme           string
		BearerFormat     string
		Flows            *OAuthFlows
		OpenIDConnectURL string
	}

	securitySchemeRenderer struct {
		Type             string              `json:"type" yaml:"type"`
		Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
		Name             string              `json:"name" yaml:"name"`
		In               string              `json:"in" yaml:"in"`
		Scheme           string              `json:"scheme" yaml:"scheme"`
		BearerFormat     string              `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
		Flows            *oauthFlowsRenderer `json:"flows" yaml:"flows"`
		OpenIDConnectURL string              `json:"openIdConnectUrl" yaml:"openIdConnectUrl"`
	}

	OAuthFlows struct {
		Implicit          *OAuthFlow
		Password          *OAuthFlow
		ClientCredentials *OAuthFlow
		AuthorizationCode *OAuthFlow
	}

	oauthFlowsRenderer struct {
		Implicit          *oauthFlowRenderer `json:"implicit,omitempty" yaml:"implicit,omitempty"`
		Password          *oauthFlowRenderer `json:"password,omitempty" yaml:"password,omitempty"`
		ClientCredentials *oauthFlowRenderer `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
		AuthorizationCode *oauthFlowRenderer `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
	}

	OAuthFlow struct {
		AuthorizationURL string
		TokenUrl         string
		RefreshUrl       string
		Scopes           map[string]web.LocaleStringer
	}

	oauthFlowRenderer struct {
		AuthorizationURL string                                 `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
		TokenUrl         string                                 `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
		RefreshUrl       string                                 `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
		Scopes           *orderedmap.OrderedMap[string, string] `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	}

	SecurityRequirement struct {
		Name   string
		Values []string
	}

	securityRequirementRenderer = orderedmap.OrderedMap[string, []string]
)

func (e *SecurityScheme) build(p *message.Printer) *securitySchemeRenderer {
	return &securitySchemeRenderer{
		Type:             e.Type,
		Description:      sprint(p, e.Description),
		Name:             sprint(p, e.Name),
		In:               e.In,
		Scheme:           e.Scheme,
		BearerFormat:     e.BearerFormat,
		Flows:            e.Flows.build(p),
		OpenIDConnectURL: e.OpenIDConnectURL,
	}
}

func (f *OAuthFlows) build(p *message.Printer) *oauthFlowsRenderer {
	return &oauthFlowsRenderer{
		Implicit:          f.Implicit.build(p),
		Password:          f.Password.build(p),
		ClientCredentials: f.ClientCredentials.build(p),
		AuthorizationCode: f.AuthorizationCode.build(p),
	}
}

func (f *OAuthFlow) build(p *message.Printer) *oauthFlowRenderer {
	return &oauthFlowRenderer{
		AuthorizationURL: f.AuthorizationURL,
		TokenUrl:         f.TokenUrl,
		RefreshUrl:       f.RefreshUrl,
		Scopes:           writeMap2OrderedMap(f.Scopes, nil, func(in web.LocaleStringer) string { return sprint(p, in) }),
	}
}
