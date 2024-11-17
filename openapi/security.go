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
		ID string // 在 components 中的键名，要求唯一性。

		Type             string
		Description      web.LocaleStringer
		Name             string
		In               string
		Scheme           string
		BearerFormat     string
		Flows            *OAuthFlows
		OpenIDConnectURL string
	}

	securitySchemeRenderer struct {
		Type             string              `json:"type" yaml:"type"`
		Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
		Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
		In               string              `json:"in,omitempty" yaml:"in,omitempty"`
		Scheme           string              `json:"scheme,omitempty" yaml:"scheme,omitempty"`
		BearerFormat     string              `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
		Flows            *oauthFlowsRenderer `json:"flows,omitempty" yaml:"flows,omitempty"`
		OpenIDConnectURL string              `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
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
		TokenURL         string
		RefreshURL       string
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
		Scopes []string
	}

	securityRequirementRenderer = orderedmap.OrderedMap[string, []string]
)

func (e *SecurityScheme) valid() *web.FieldError {
	switch e.Type {
	case SecuritySchemeTypeAPIKey:
		if e.Name == "" {
			return web.NewFieldError("Name", "不能为空")
		}
		if e.In != InQuery && e.In != InHeader && e.In != InCookie {
			return web.NewFieldError("In", "无效的值")
		}
	case SecuritySchemeTypeHTTP:
		if e.Scheme == "" {
			return web.NewFieldError("Scheme", "不能为空")
		}
	case SecuritySchemeTypeOAuth2:
		if e.Flows == nil {
			return web.NewFieldError("Flows", "不能为空")
		}
	case SecuritySchemeTypeOpenIDConnect:
		if e.OpenIDConnectURL == "" {
			return web.NewFieldError("OpenIDConnectURL", "不能为空")
		}
	case SecuritySchemeTypeMutualTLS:
	default:
		return web.NewFieldError("Type", "无效的值")
	}
	return nil
}

func (e *SecurityScheme) build(p *message.Printer) *securitySchemeRenderer {
	return &securitySchemeRenderer{
		// NOTE: SecurityScheme.ID 作为 components.securitySchemes 的键名使用，并不出现在 securitySchemeRenderer
		Type:             e.Type,
		Description:      sprint(p, e.Description),
		Name:             e.Name,
		In:               e.In,
		Scheme:           e.Scheme,
		BearerFormat:     e.BearerFormat,
		Flows:            e.Flows.build(p),
		OpenIDConnectURL: e.OpenIDConnectURL,
	}
}

func (f *OAuthFlows) build(p *message.Printer) *oauthFlowsRenderer {
	if f == nil {
		return nil
	}

	return &oauthFlowsRenderer{
		Implicit:          f.Implicit.build(p),
		Password:          f.Password.build(p),
		ClientCredentials: f.ClientCredentials.build(p),
		AuthorizationCode: f.AuthorizationCode.build(p),
	}
}

func (f *OAuthFlow) build(p *message.Printer) *oauthFlowRenderer {
	if f == nil {
		return nil
	}

	return &oauthFlowRenderer{
		AuthorizationURL: f.AuthorizationURL,
		TokenUrl:         f.TokenURL,
		RefreshUrl:       f.RefreshURL,
		Scopes:           writeMap2OrderedMap(f.Scopes, nil, func(in web.LocaleStringer) string { return sprint(p, in) }),
	}
}
