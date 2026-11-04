// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"slices"
	"strconv"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/filter/validators"
	"github.com/issue9/web/locales"
)

func createFilter() (filter.Map, *filter.Filter) {
	m, c := filter.CreateMapCollector()
	f := filter.New("", true, c)
	return m, f
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (req *Request) valid(f *filter.Filter, skipRefNotNil bool) {
	if skipRefNotNil && req.Ref != nil {
		return
	}

	if req.Body != nil {
		f.ApplyFilter("Body", func(f *filter.Filter) { req.Body.valid(f, skipRefNotNil) })
	}

	for k, v := range req.Content {
		f.ApplyFilter("Content["+k+"]", func(f *filter.Filter) { v.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (resp *Response) valid(f *filter.Filter, skipRefNotNil bool) {
	if skipRefNotNil && resp.Ref != nil {
		return
	}

	for i, p := range resp.Headers {
		f.ApplyFilter("Headers["+strconv.Itoa(i)+"]", func(f *filter.Filter) { p.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}

	if resp.Body != nil {
		f.ApplyFilter("Body", func(f *filter.Filter) { resp.Body.valid(f, skipRefNotNil) })
	}

	for k, v := range resp.Content {
		f.ApplyFilter("Content["+k+"]", func(f *filter.Filter) { v.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}
}

func (s *Server) valid(f *filter.Filter) {
	params := getPathParams(s.URL)
	eq := func(v []*ServerVariable) bool { return len(params) == len(s.Variables) }
	containe := func(v string) bool { return slices.Contains(params, v) }

	f.Valid("Variables", s.Variables, eq, localeutil.Phrase("The size of parameters in the URL not equal the size of variables"))

	for _, v := range s.Variables {
		f.Valid("Variables", v.Name, containe, localeutil.Phrase("The parameters %s not in path", v.Name))
	}
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (t *Parameter) valid(f *filter.Filter, skipRefNotNil bool) {
	if skipRefNotNil && t.Ref != nil {
		return
	}

	f.Valid("Name", t.Name, validators.Not(validators.Empty), locales.CanNotBeEmpty).
		Valid("Schema", t.Schema, func(v *Schema) bool { return v != nil }, locales.CanNotBeEmpty).
		When(func() bool { return t.Schema != nil }, func(v *filter.Filter) {
			v.ApplyFilter("Schema", func(f *filter.Filter) { t.Schema.valid(f, skipRefNotNil) }).
				Valid("Schema", t.Schema.isBasicType(), func(v bool) bool { return v }, localeutil.Phrase("Only basic type are supported"))
		})
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (s *Schema) valid(f *filter.Filter, skipRefNotNil bool) {
	if skipRefNotNil && s.Ref != nil {
		return
	}

	if s.Type == "" && len(s.AnyOf) == 0 && len(s.AllOf) == 0 && len(s.OneOf) == 0 {
		f.AddReason("Type", locales.CanNotBeEmpty)
		return
	}

	for i, item := range s.AllOf {
		f.ApplyFilter("AllOf["+strconv.Itoa(i)+"]", func(f *filter.Filter) { item.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}

	for i, item := range s.AnyOf {
		f.ApplyFilter("AnyOf["+strconv.Itoa(i)+"]", func(f *filter.Filter) { item.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}

	for i, item := range s.OneOf {
		f.ApplyFilter("OneOf["+strconv.Itoa(i)+"]", func(f *filter.Filter) { item.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}

	if s.Items != nil {
		f.ApplyFilter("Items", func(f *filter.Filter) { s.Items.valid(f, skipRefNotNil) }).
			Valid("Type", s.Type, func(t string) bool { return t == TypeArray }, localeutil.Phrase("The value must be %s", TypeArray))
	}
	if s.Type == TypeArray && s.Items == nil {
		f.AddReason("Type", localeutil.Phrase("The value must be %s", TypeArray))
		return
	}

	for key, item := range s.Properties {
		f.ApplyFilter("Properties["+key+"]", func(f *filter.Filter) { item.valid(f, skipRefNotNil) })
		if f.HasError() {
			return
		}
	}
	if len(s.Properties) > 0 && s.Type != TypeObject {
		f.AddReason("Type", localeutil.Phrase("The value must be %s", TypeObject))
		return
	}

	if s.AdditionalProperties != nil {
		f.ApplyFilter("AdditionalProperties", func(f *filter.Filter) { s.AdditionalProperties.valid(f, skipRefNotNil) })
	}
}

func (e *SecurityScheme) valid(f *filter.Filter) {
	switch e.Type {
	case SecuritySchemeTypeAPIKey:
		f.Valid("Name", e.Name, validators.Not(validators.Empty), locales.CanNotBeEmpty).
			Valid("In", e.In, func(in string) bool { return e.In == InQuery || e.In == InHeader || e.In == InCookie }, locales.InvalidValue)
	case SecuritySchemeTypeHTTP:
		f.Valid("Scheme", e.Scheme, validators.Not(validators.Empty), locales.CanNotBeEmpty)
	case SecuritySchemeTypeOAuth2:
		f.Valid("Flows", e.Flows, func(f *OAuthFlows) bool { return f != nil }, locales.CanNotBeEmpty)
	case SecuritySchemeTypeOpenIDConnect:
		f.Valid("OpenIDConnectURL", e.OpenIDConnectURL, validators.Not(validators.Empty), locales.CanNotBeEmpty)
	case SecuritySchemeTypeMutualTLS:
	default:
		f.AddReason("Type", locales.InvalidValue)
	}
}
