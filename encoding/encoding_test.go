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

	a.Equal(len(marshals), 1)
	a.Nil(Marshal("not exists"))
	a.NotNil(Marshal(DefaultEncoding))

	// 添加已存在的
	a.Equal(AddMarshal(DefaultEncoding, json.Marshal), ErrExists)
	a.Equal(len(marshals), 1) // 添加没成功

	a.NotError(AddMarshal("json", json.Marshal))
	a.Equal(len(marshals), 2) // 添加没成功
	a.NotNil(Marshal("json"))
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.Equal(len(unmarshals), 1)
	a.Nil(Unmarshal("not exists"))
	a.NotNil(Unmarshal(DefaultEncoding))

	// 添加已存在的
	a.Equal(AddUnmarshal(DefaultEncoding, json.Unmarshal), ErrExists)
	a.Equal(len(unmarshals), 1) // 添加没成功

	a.NotError(AddUnmarshal("json", json.Unmarshal))
	a.Equal(len(unmarshals), 2) // 添加没成功
	a.NotNil(Unmarshal("json"))
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t)

	a.Equal("application/xml; charset=utf16", BuildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+DefaultCharset, BuildContentType("application/xml", ""))
	a.Equal(DefaultEncoding+"; charset="+DefaultCharset, BuildContentType("", ""))
	a.Equal("application/xml; charset="+DefaultCharset, BuildContentType("application/xml", ""))
}

func TestParseContentType(t *testing.T) {
	a := assert.New(t)

	e, c := ParseContentType("")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = ParseContentType(" ")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = ParseContentType(" ;;;")
	a.Equal(e, DefaultEncoding).Equal(c, DefaultCharset)

	e, c = ParseContentType("application/XML")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = ParseContentType("application/XML;")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = ParseContentType(" application/xml ;;")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = ParseContentType("  ;charset=utf16")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")

	e, c = ParseContentType("  ;charset=utf16;")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")

	e, c = ParseContentType("application/xml; charset=utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = ParseContentType("application/xml;utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = ParseContentType("application/xml;=utf16")
	a.Equal(e, "application/xml").Equal(c, "utf16")

	e, c = ParseContentType("application/xml;utf16=")
	a.Equal(e, "application/xml").Equal(c, DefaultCharset)

	e, c = ParseContentType("application/xml;=utf16=")
	a.Equal(e, "application/xml").Equal(c, "utf16=")

	e, c = ParseContentType(";charset=utf16")
	a.Equal(e, DefaultEncoding).Equal(c, "utf16")
}
