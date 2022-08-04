// SPDX-License-Identifier: MIT

package header

import (
	"testing"

	"github.com/issue9/assert/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestParseWithParam(t *testing.T) {
	a := assert.New(t, false)

	v, p := ParseWithParam(";;;", "charset")
	a.Equal(v, "").Empty(p)

	v, p = ParseWithParam("application/xml;charset=utf-8", "charset")
	a.Equal(v, "application/xml").Equal(p, "utf-8")

	// charset=UTF-8
	v, p = ParseWithParam("application/xml;\tCHARSet=UTF-8;p1=k1;p2=k2", "charset")
	a.Equal(v, "application/xml").Equal(p, "utf-8")

	// pk2=k2;pk3;charset="UTF-8" // 没有值的参数
	v, p = ParseWithParam(`application/xml;p1=k1;p2=k2;pk3;CHARSet="UTF-8"`, "charset")
	a.Equal(v, "application/xml").Equal(p, "utf-8")

	// pk2=k2;pk3;charset // 没有值的参数
	v, p = ParseWithParam(`application/xml;p1=k1;p2=k2;pk3;charset`, "charset")
	a.Equal(v, "application/xml").Equal(p, "")

	// charset=
	v, p = ParseWithParam("application/xml; charset=", "charset")
	a.Equal(v, "application/xml").Equal(p, "")

	// 没有 charset
	v, p = ParseWithParam("application/xml;", "charset")
	a.Equal(v, "application/xml").Equal(p, "")

	// 没有 ;charset
	v, p = ParseWithParam("application/xml", "charset")
	a.Equal(v, "application/xml").Equal(p, "")

	// 参数格式不正确
	v, p = ParseWithParam("application/xml; invalid-params", "charset")
	a.Equal(v, "application/xml").Equal(p, "")
}

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t, false)

	name, enc := AcceptCharset(UTF8Name)
	a.Equal(name, UTF8Name).
		True(CharsetIsNop(enc))

	name, enc = AcceptCharset("")
	a.Equal(name, UTF8Name).
		True(CharsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc = AcceptCharset("*")
	a.Equal(name, UTF8Name).
		True(CharsetIsNop(enc))

	name, enc = AcceptCharset("gbk")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 传递一个非正规名称
	name, enc = AcceptCharset("chinese")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// q 错解析错误
	name, enc = AcceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 不支持的编码
	name, enc = AcceptCharset("not-supported")
	a.Empty(name).
		Nil(enc)
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t, false)

	a.Equal("application/xml; charset=utf16", BuildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+UTF8Name, BuildContentType("application/xml", ""))
	a.Equal("application/xml; charset="+UTF8Name, BuildContentType("application/xml", ""))
	a.PanicString(func() {
		BuildContentType("", "")
	}, "mt 不能为空")
}
