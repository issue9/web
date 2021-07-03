// SPDX-License-Identifier: MIT

package content

import (
	"testing"

	"github.com/issue9/assert"
)

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c, err := ParseContentType("")
	a.NotError(err).Equal(e, DefaultMimetype).Equal(c, DefaultCharset)

	e, c, err = ParseContentType(" ")
	a.NotError(err).Equal(e, DefaultMimetype).Equal(c, DefaultCharset)

	_, _, err = ParseContentType(";charset=utf-8")
	a.Error(err)

	e, c, err = ParseContentType(" ;;;")
	a.Error(err).Empty(e).Empty(c)

	e, c, err = ParseContentType("application/XML")
	a.NotError(err).Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c, err = ParseContentType("application/XML;")
	a.NotError(err).Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c, err = ParseContentType("text/html;charset=utF-8")
	a.NotError(err).Equal(e, "text/html").Equal(c, "utF-8")

	e, c, err = ParseContentType(`Text/HTML;Charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, "gbk")

	e, c, err = ParseContentType(`Text/HTML; charset="Gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, "Gbk")

	e, c, err = ParseContentType(`multipart/form-data; boundary=AaB03x`)
	a.NotError(err).Equal(e, "multipart/form-data").Equal(c, DefaultCharset)
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml; charset=utf16", BuildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+DefaultCharset, BuildContentType("application/xml", ""))
	a.Equal(DefaultMimetype+"; charset="+DefaultCharset, BuildContentType("", ""))
	a.Equal("application/xml; charset="+DefaultCharset, BuildContentType("application/xml", ""))
}
