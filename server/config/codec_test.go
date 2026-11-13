// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

func TestConfigOf_buildCodec(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{
		Compressors: []*compressConfig{
			{Types: []string{"text/*", "application/*"}, ID: "compress-msb-8"},
			{Types: []string{"text/*"}, ID: "br-default"},
			{Types: []string{"application/*"}, ID: "gzip-default"},
		},
	}
	e, f := buildFilter()
	conf.buildCodec(f)
	a.False(f.HasError()).Empty(e)

	conf = &configOf[empty]{
		Compressors: []*compressConfig{
			{Types: []string{"text/*"}, ID: "compress-msb-8"},
			{Types: []string{"text/*"}, ID: "not-exists-id"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).NotEmpty(e["compresses[1].id"])
}

func TestConfigOf_sanitizeMediaTypes(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json"},
			{Target: "xml", Accept: "request,response"},
		},
	}
	e, f := buildFilter()
	conf.buildCodec(f)
	a.False(f.HasError()).NotNil(conf.codec).Empty(e)

	conf = &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json"},
			{Target: "not-exists"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e, filter.Map{"mediaTypes[1].target": []web.LocaleStringer{locales.NotFound}})

	// Accept

	conf = &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json", Accept: "req"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e, filter.Map{"mediaTypes[0].accept": []web.LocaleStringer{locales.InvalidValue}})

	conf = &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json", Accept: "request,resp"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e, filter.Map{"mediaTypes[0].accept": []web.LocaleStringer{locales.InvalidValue}})

	conf = &configOf[empty]{
		MediaTypes: []*mediaTypeConfig{
			{Target: "json", Accept: "request,response,request"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e, filter.Map{"mediaTypes[0].accept": []web.LocaleStringer{locales.InvalidValue}})
}

func TestRegisterMediaType(t *testing.T) {
	a := assert.New(t, false)

	v, f := mediaTypeFactory.get("json")
	a.True(f).NotNil(v)

	a.PanicString(func() {
		RegisterMediaType(codec.NewJSON(), "json2")
	}, "存在同名的媒体类型")

	RegisterMediaType(codec.ForkMediaType(codec.NewJSON(), "application/jsonv2", "application/problem+jsonv2"), "json2")
	v, f = mediaTypeFactory.get("json2")
	a.True(f).NotNil(v)
}
