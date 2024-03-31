// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package qheader

import (
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"
	"github.com/issue9/mux/v8/header"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestParseWithParam(t *testing.T) {
	a := assert.New(t, false)

	v, p := ParseWithParam(";;;", "charset")
	a.Equal(v, "").Empty(p)

	v, p = ParseWithParam("application/xml;charset=utf-8", "charset")
	a.Equal(v, "application/xml").Equal(p, "utf-8")

	v, p = ParseWithParam("application/xml;charset=utf-8", "")
	a.Equal(v, "application/xml").Equal(p, "")

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

	name, enc := ParseAcceptCharset(header.UTF8)
	a.Equal(name, header.UTF8).
		True(CharsetIsNop(enc))

	name, enc = ParseAcceptCharset("")
	a.Equal(name, header.UTF8).
		True(CharsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc = ParseAcceptCharset("*")
	a.Equal(name, header.UTF8).
		True(CharsetIsNop(enc))

	name, enc = ParseAcceptCharset("gbk")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 传递一个非正规名称
	name, enc = ParseAcceptCharset("chinese")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// q 错解析错误
	name, enc = ParseAcceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 不支持的编码
	name, enc = ParseAcceptCharset("not-supported")
	a.Empty(name).
		Nil(enc)
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t, false)

	a.Equal("application/xml; charset=utf16", BuildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+header.UTF8, BuildContentType("application/xml", ""))
	a.Equal("application/xml; charset="+header.UTF8, BuildContentType("application/xml", ""))
	a.PanicString(func() {
		BuildContentType("", "")
	}, "mt 不能为空")
}

func TestClientIP(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Post(a, "/path", nil).Request()
	r.RemoteAddr = "192.168.1.1"
	a.Equal(ClientIP(r), r.RemoteAddr)

	r = rest.Post(a, "/path", nil).Header("x-real-ip", "192.168.1.1:8080").Request()
	a.Equal(ClientIP(r), "192.168.1.1:8080")

	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Request()
	a.Equal(ClientIP(r), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	a.Equal(ClientIP(r), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("Remote-Addr", "192.168.2.0").
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	a.Equal(ClientIP(r), "192.168.2.1:8080")
}

func TestInitETag(t *testing.T) {
	a := assert.New(t, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).Request()
	a.False(InitETag(w, r, `"abc"`, false)).
		Equal(w.Header().Get(header.ETag), `"abc"`)

	// client=abc, etag=abc, weak=false
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).Header(header.IfNoneMatch, `"abc"`).Request()
	a.True(InitETag(w, r, `"abc"`, false)).
		Equal(w.Header().Get(header.ETag), `"abc"`)

	// client=abc, etag=abc, weak=true
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).Header(header.IfNoneMatch, `"abc"`).Request()
	a.True(InitETag(w, r, `"abc"`, true)).
		Equal(w.Header().Get(header.ETag), `W/"abc"`)

	// client=W/abc, etag=abc, weak=true
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).Header(header.IfNoneMatch, `W/"abc"`).Request()
	a.True(InitETag(w, r, `"abc"`, true)).
		Equal(w.Header().Get(header.ETag), `W/"abc"`)

	// client=abcdef, etag=abc, weak=true
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).Header(header.IfNoneMatch, `"abcdef"`).Request()
	a.False(InitETag(w, r, `"abc"`, true)).
		Equal(w.Header().Get(header.ETag), `W/"abc"`)

	// client=abcdef, etag=abc, weak=false
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).Header(header.IfNoneMatch, `"abcdef"`).Request()
	a.False(InitETag(w, r, `"abc"`, false)).
		Equal(w.Header().Get(header.ETag), `"abc"`)
}
