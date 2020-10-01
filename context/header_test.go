// SPDX-License-Identifier: MIT

package context

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"

	"github.com/issue9/web/context/mimetype"
)

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	name, enc, err := acceptCharset(utfName)
	a.NotError(err).
		Equal(name, utfName).
		True(charsetIsNop(enc))

	name, enc, err = acceptCharset("")
	a.NotError(err).
		Equal(name, utfName).
		True(charsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc, err = acceptCharset("*")
	a.NotError(err).
		Equal(name, utfName).
		True(charsetIsNop(enc))

	name, enc, err = acceptCharset("gbk")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 传递一个非正规名称
	name, enc, err = acceptCharset("chinese")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// q 错解析错误
	name, enc, err = acceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 不支持的编码
	name, enc, err = acceptCharset("not-supported")
	a.Error(err).
		Empty(name).
		Nil(enc)
}

func TestServer_acceptLanguage(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	tag := srv.acceptLanguage("")
	a.Equal(tag, language.Und)

	tag = srv.acceptLanguage("zh")
	a.Equal(tag, language.Chinese)

	tag = srv.acceptLanguage("zh-Hant")
	a.Equal(tag, language.TraditionalChinese)

	tag = srv.acceptLanguage("zh-Hans")
	a.Equal(tag, language.SimplifiedChinese)

	tag = srv.acceptLanguage("zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.Equal(tag, language.English)
}

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c, err := parseContentType("")
	a.NotError(err).Equal(e, mimetype.DefaultMimetype).Equal(c, utfName)

	e, c, err = parseContentType(" ")
	a.NotError(err).Equal(e, mimetype.DefaultMimetype).Equal(c, utfName)

	e, c, err = parseContentType(" ;;;")
	a.Error(err).Empty(e).Empty(c)

	e, c, err = parseContentType("application/XML")
	a.NotError(err).Equal(e, "application/xml").Equal(c, utfName)

	e, c, err = parseContentType("application/XML;")
	a.NotError(err).Equal(e, "application/xml").Equal(c, utfName)

	e, c, err = parseContentType("text/html;charset=utf-8")
	a.NotError(err).Equal(e, "text/html").Equal(c, "utf-8")

	e, c, err = parseContentType(`Text/HTML;Charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, utfName)

	e, c, err = parseContentType(`Text/HTML; charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, "gbk")

	e, c, err = parseContentType(`multipart/form-data; boundary=AaB03x`)
	a.NotError(err).Equal(e, "multipart/form-data").Equal(c, utfName)
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml; charset=utf16", buildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+utfName, buildContentType("application/xml", ""))
	a.Equal(mimetype.DefaultMimetype+"; charset="+utfName, buildContentType("", ""))
	a.Equal("application/xml; charset="+utfName, buildContentType("application/xml", ""))
}
