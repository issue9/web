// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"

	"github.com/issue9/web/encoding"
)

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	name, enc, err := acceptCharset(encoding.DefaultCharset)
	a.NotError(err).
		Equal(name, encoding.DefaultCharset).
		True(charsetIsNop(enc))

	name, enc, err = acceptCharset("")
	a.NotError(err).
		Equal(name, encoding.DefaultCharset).
		True(charsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc, err = acceptCharset("*")
	a.NotError(err).
		Equal(name, encoding.DefaultCharset).
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
	a.Error(err).
		Equal(name, "").
		Nil(enc)

	// 不支持的编码
	name, enc, err = acceptCharset("not-supported")
	a.Error(err).
		Empty(name).
		Nil(enc)
}

func TestAcceptLanguage(t *testing.T) {
	a := assert.New(t)
	tag, err := acceptLanguage("")
	a.NotError(err).Equal(tag, language.Und)

	tag, err = acceptLanguage("xx;q=xxx")
	a.Error(err).Equal(tag, language.Und)

	tag, err = acceptLanguage("zh")
	a.NotError(err).Equal(tag, language.Chinese)

	tag, err = acceptLanguage("zh-Hant")
	a.NotError(err).Equal(tag, language.TraditionalChinese)

	tag, err = acceptLanguage("zh-Hans")
	a.NotError(err).Equal(tag, language.SimplifiedChinese)

	tag, err = acceptLanguage("zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.NotError(err).Equal(tag, language.English)
}

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c, err := parseContentType("")
	a.NotError(err).Equal(e, encoding.DefaultMimeType).Equal(c, encoding.DefaultCharset)

	e, c, err = parseContentType(" ")
	a.NotError(err).Equal(e, encoding.DefaultMimeType).Equal(c, encoding.DefaultCharset)

	e, c, err = parseContentType(" ;;;")
	a.Error(err).Empty(e).Empty(c)

	e, c, err = parseContentType("application/XML")
	a.NotError(err).Equal(e, "application/xml").Equal(c, encoding.DefaultCharset)

	e, c, err = parseContentType("application/XML;")
	a.NotError(err).Equal(e, "application/xml").Equal(c, encoding.DefaultCharset)

	e, c, err = parseContentType("text/html;charset=utf-8")
	a.NotError(err).Equal(e, "text/html").Equal(c, "utf-8")

	e, c, err = parseContentType(`Text/HTML;Charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, encoding.DefaultCharset)

	e, c, err = parseContentType(`Text/HTML; charset="gbk"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, "gbk")

	e, c, err = parseContentType(`multipart/form-data; boundary=AaB03x`)
	a.NotError(err).Equal(e, "multipart/form-data").Equal(c, encoding.DefaultCharset)
}
