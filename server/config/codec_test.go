// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v4"

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
	a.NotError(conf.buildCodec())

	conf = &configOf[empty]{
		Compressors: []*compressConfig{
			{Types: []string{"text/*"}, ID: "compress-msb-8"},
			{Types: []string{"text/*"}, ID: "not-exists-id"},
		},
	}
	err := conf.buildCodec()
	a.Error(err).Equal(err.Field, "compresses[1].id")
}

func TestConfigOf_sanitizeMimetypes(t *testing.T) {
	a := assert.New(t, false)

	conf := &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json"},
			{Type: "xml", Target: "xml"},
		},
	}
	a.NotError(conf.buildCodec())

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json"},
			{Type: "json", Target: "xml"},
		},
	}
	err := conf.buildCodec()
	a.Error(err).Equal(err.Field, "mimetypes[1].type")

	conf = &configOf[empty]{
		Mimetypes: []*mimetypeConfig{
			{Type: "json", Target: "json"},
			{Type: "xml", Target: "not-exists"},
		},
	}
	err = conf.buildCodec()
	a.Error(err).
		Equal(err.Field, "mimetypes[1].target").
		Equal(err.Message, locales.ErrNotFound())
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
