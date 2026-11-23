// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

func TestConfigOf_buildNegotiation(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{
		Compressors: []*compressConfig{
			{Types: []string{"text/*", "application/*"}, ID: "compress-msb-8"},
			{Types: []string{"text/*"}, ID: "br-default"},
			{Types: []string{"application/*"}, ID: "gzip-default"},
		},
	}
	e, f := buildFilter()
	conf.buildNegotiation(f)
	a.False(f.HasError()).Empty(e)

	conf = &configOf[empty]{
		Compressors: []*compressConfig{
			{Types: []string{"text/*"}, ID: "compress-msb-8"},
			{Types: []string{"text/*"}, ID: "not-exists-id"},
		},
	}
	e, f = buildFilter()
	conf.buildNegotiation(f)
	a.True(f.HasError()).NotEmpty(e["compresses[1].id"])
}

func TestConfigOf_sanitizeMediaTypes(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json", Name: "application/json"},
			{Target: "xml", Name: "application/xml"},
		},
	}
	e, f := buildFilter()
	conf.buildNegotiation(f)
	a.False(f.HasError()).NotNil(conf.negotiation).Empty(e)

	conf = &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json", Name: "application/json"},
			{Target: "not-exists", Name: "application/not-exists"},
		},
	}
	e, f = buildFilter()
	conf.buildNegotiation(f)
	a.True(f.HasError()).
		Equal(e, filter.Map{"mediaTypes[1].target": []web.LocaleStringer{locales.NotFound}})
}
