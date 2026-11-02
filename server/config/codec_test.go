// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/locales"
	"github.com/issue9/web/mimetype/json"
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
	a.False(f.HasError()).Empty(e.Name)

	conf = &configOf[empty]{
		Compressors: []*compressConfig{
			{Types: []string{"text/*"}, ID: "compress-msb-8"},
			{Types: []string{"text/*"}, ID: "not-exists-id"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).Equal(e.Name, "compresses[1].id")
}

func TestConfigOf_sanitizeMimetypes(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json"},
			{Type: "xml", Target: "xml", Accept: "request,response"},
		},
	}
	e, f := buildFilter()
	conf.buildCodec(f)
	a.False(f.HasError()).NotNil(conf.codec).Empty(e.Name)

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json"},
			{Type: "json", Target: "xml"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).Equal(e.Name, "mimetypes[1].type")

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json"},
			{Type: "xml", Target: "not-exists"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e.Name, "mimetypes[1].target").
		Equal(e.Message, locales.NotFound)

	// Accept

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json", Accept: "req"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e.Name, "mimetypes[0].accept").
		Equal(e.Message, locales.InvalidValue)

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json", Accept: "request,resp"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e.Name, "mimetypes[0].accept").
		Equal(e.Message, locales.InvalidValue)

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json", Accept: "request,response,request"},
		},
	}
	e, f = buildFilter()
	conf.buildCodec(f)
	a.True(f.HasError()).
		Equal(e.Name, "mimetypes[0].accept").
		Equal(e.Message, locales.InvalidValue)
}

func TestRegisterMimetype(t *testing.T) {
	a := assert.New(t, false)

	v, f := mimetypesFactory.get("json")
	a.True(f).
		NotNil(v).
		NotNil(v.marshal)

	RegisterMimetype(json.Marshal, json.Unmarshal, "json")
	v, f = mimetypesFactory.get("json")
	a.True(f).
		NotNil(v).
		NotNil(v.marshal)
}
