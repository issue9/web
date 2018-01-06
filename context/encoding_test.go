// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestAddMarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(v interface{}) ([]byte, error) { return nil, nil }

	a.Equal(0, len(marshals))

	a.NotError(AddMarshal("n1", f1))
	a.NotError(AddMarshal("n2", f1))
	a.Equal(AddMarshal("n2", f1), ErrExists)
	a.Equal(2, len(marshals))
}

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(data []byte, v interface{}) error { return nil }

	a.Equal(0, len(unmarshals))

	a.NotError(AddUnmarshal("n1", f1))
	a.NotError(AddUnmarshal("n2", f1))
	a.Equal(AddUnmarshal("n2", f1), ErrExists)
	a.Equal(2, len(unmarshals))
}

func TestAddCharset(t *testing.T) {
	a := assert.New(t)
	e := simplifiedchinese.GBK

	a.Equal(1, len(charset))
	a.Nil(charset[DefaultCharset])

	a.NotError(AddCharset("n1", e))
	a.NotError(AddCharset("n2", e))
	a.Equal(AddCharset("n2", e), ErrExists)
	a.Equal(3, len(charset))
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml;charset=utf16", buildContentType("application/xml", "utf16"))
	a.Equal("application/xml;charset="+DefaultCharset, buildContentType("application/xml", ""))
	a.Equal("application/json;charset=utf-8", buildContentType("", ""))
	a.Equal("application/xml;charset=utf-8", buildContentType("application/xml", ""))
}

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c := parseContentType("")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = parseContentType(" ")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = parseContentType(" ;;;")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = parseContentType("application/XML")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType("application/XML;")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType(" application/xml ;;")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType("  ;charset=utf16")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")

	e, c = parseContentType("  ;charset=utf16;")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")

	e, c = parseContentType("application/xml;charset=utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = parseContentType("application/xml;utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = parseContentType("application/xml;=utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = parseContentType("application/xml;utf16=")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = parseContentType("application/xml;=utf16=")
	a.Equal(e, "application/xml").Equal(c, "utf16=")

	e, c = parseContentType(";charset=utf16")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")
}
