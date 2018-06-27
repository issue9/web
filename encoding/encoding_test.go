// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding/gob"
	xencoding "golang.org/x/text/encoding"
)

func TestContentType(t *testing.T) {
	a := assert.New(t)

	um, c, err := ContentType(BuildContentType("", ""))
	a.Error(err).
		Nil(um).
		Nil(c)

	a.NotError(AddUnmarshal(DefaultMimeType, gob.Unmarshal))
	a.NotError(AddMarshal(DefaultMimeType, gob.Marshal))
	um, c, err = ContentType(BuildContentType("", ""))
	a.NotError(err).
		Equal(um, UnmarshalFunc(um)).
		Equal(c, xencoding.Nop)

	// 未指定 memetype
	um, c, err = ContentType(";" + DefaultCharset)
	a.Error(err).Nil(um).Nil(c)

	// mimetype 无法找到
	um, c, err = ContentType(BuildContentType("not-exists", DefaultCharset))
	a.Error(err).Nil(um).Nil(c)

	// charset 无法找到
	um, c, err = ContentType(BuildContentType(DefaultMimeType, "not-exists"))
	a.Error(err).Nil(um).Nil(c)
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml; charset=utf16", BuildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+DefaultCharset, BuildContentType("application/xml", ""))
	a.Equal(DefaultMimeType+"; charset="+DefaultCharset, BuildContentType("", ""))
	a.Equal("application/xml; charset="+DefaultCharset, BuildContentType("application/xml", ""))
}

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c, err := ParseContentType("")
	a.NotError(err).Equal(e, DefaultMimeType).Equal(c, DefaultCharset)

	e, c, err = ParseContentType(" ")
	a.NotError(err).Equal(e, DefaultMimeType).Equal(c, DefaultCharset)

	e, c, err = ParseContentType(" ;;;")
	a.Error(err).Empty(e).Empty(c)

	e, c, err = ParseContentType("application/XML")
	a.NotError(err).Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c, err = ParseContentType("application/XML;")
	a.NotError(err).Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c, err = ParseContentType("text/html;charset=utf-8")
	a.NotError(err).Equal(e, "text/html").Equal(c, "utf-8")

	e, c, err = ParseContentType(`Text/HTML;Charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, DefaultCharset)

	e, c, err = ParseContentType(`Text/HTML; charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, "gbk")
}
