// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestCharset(t *testing.T) {
	a := assert.New(t)

	a.Equal(len(charset), 1) // 有一条默认的字符集信息
	a.Nil(Charset("not exists"))
	a.NotNil(Charset(DefaultCharset))

	// 添加已存在的
	a.Equal(AddCharset(DefaultCharset, simplifiedchinese.GBK), ErrExists)
	a.Equal(len(charset), 1) // 添加没成功

	a.NotError(AddCharset("GBK", simplifiedchinese.GBK))
	a.Equal(len(charset), 2) // 添加没成功
	a.NotNil(Charset("GBK"))
}

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	a.Nil(Marshal("not exists"))
	a.NotNil(Marshal(DefaultMimeType))

	// 添加已存在的
	a.Equal(AddMarshal(DefaultMimeType, json.Marshal), ErrExists)

	a.NotError(AddMarshal("json", json.Marshal))
	a.NotNil(Marshal("json"))
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.Nil(Unmarshal("not exists"))
	a.NotNil(Unmarshal(DefaultMimeType))

	// 添加已存在的
	a.Equal(AddUnmarshal(DefaultMimeType, json.Unmarshal), ErrExists)

	a.NotError(AddUnmarshal("json", json.Unmarshal))
	a.NotNil(Unmarshal("json"))
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

	e, c, err = ParseContentType(`Text/HTML;Charset="utf-8"`)
	a.NotError(err).Equal(e, "text/html").Equal(c, "utf-8")
}
