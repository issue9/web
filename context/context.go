// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package context 对 HTTP 请求和输出作了简单的封装。
package context

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	xencoding "golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/issue9/web/encoding"
)

// Context 是对当前请求内容的封装，仅与当前请求相关。
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的媒体类型，以及名称
	OutputMimeType     encoding.MarshalFunc
	OutputMimeTypeName string

	// 输出到客户端的字符集
	//
	// 若值为 xencoding.Nop 或是空，表示为 utf-8
	OutputCharset     xencoding.Encoding
	OutputCharsetName string

	// 客户端内容所使用的媒体类型。
	InputMimeType encoding.UnmarshalFunc

	// 客户端内容所使用的字符集
	//
	// 若值为 xencoding.Nop 或是空，表示为 utf-8
	InputCharset xencoding.Encoding

	// 从客户端获取的内容，已经解析为 utf-8 方式。
	body []byte
}

// Body 获取用户提交的内容。
//
// 相对于 ctx.Request().Body，此函数可多次读取。
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.body != nil {
		return ctx.body, nil
	}

	if ctx.body, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		return nil, err
	}

	if ctx.InputCharset == nil {
		return ctx.body, nil
	}

	reader := transform.NewReader(bytes.NewReader(ctx.body), ctx.InputCharset.NewDecoder())
	ctx.body, err = ioutil.ReadAll(reader)
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象。
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	return ctx.InputMimeType(body, v)
}

// Marshal 将 v 发送给客户端。
//
// NOTE: 若在 headers 中包含了 Content-Type，则会覆盖原来的 Content-Type 报头，
// 但是不会改变其输出时的实际编码方式。
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	key := http.CanonicalHeaderKey("Content-type")
	found := false
	for k, v := range headers {
		// strings.ToLower 的性能未必有 http.CanonicalHeaderKey 好，
		// 所以直接使用了 http.CanonicalHeaderKey 作转换。
		if http.CanonicalHeaderKey(k) == key {
			found = true
		}
		ctx.Response.Header().Set(k, v)

	}
	if !found {
		ct := encoding.BuildContentType(ctx.OutputMimeTypeName, ctx.OutputCharsetName)
		ctx.Response.Header().Set("Content-Type", ct)
	}

	data, err := ctx.OutputMimeType(v)
	if err != nil {
		return err
	}

	ctx.Response.WriteHeader(status)

	if ctx.OutputCharset == nil {
		_, err = ctx.Response.Write(data)
		return err
	}

	w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
	if _, err = w.Write(data); err != nil {
		return err
	}
	return w.Close()
}

// Read 从客户端读取数据并转换成 v 对象。
//
// 功能与 Unmarshal() 相同，只不过 Read() 在出错时，
// 会直接调用 Error() 处理：输出 422 的状态码，
// 并返回一个 false，告知用户转换失败。
func (ctx *Context) Read(v interface{}) (ok bool) {
	if err := ctx.Unmarshal(v); err != nil {
		ctx.Error(http.StatusUnprocessableEntity, err)
		return false
	}

	return true
}

// Render 将 v 渲染给客户端。
//
// 功能与 Marshal() 相同，只不过 Render() 在出错时，
// 会直接调用 Error() 处理，输出 500 的状态码。
//
// 如果需要具体控制出错后的处理方式，可以使用 Marshal 函数。
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	if err := ctx.Marshal(status, v, headers); err != nil {
		ctx.Error(http.StatusInternalServerError, err)
	}
}

// ClientIP 返回客户端的 IP 地址。
//
// 获取顺序如下：
//  - X-Forwarded-For 的第一个元素
//  - Remote-Addr 报头
//  - X-Read-IP 报头
func (ctx *Context) ClientIP() string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(ip, ','); index > 0 {
		ip = ip[:index]
	}
	if ip == "" {
		if ctx.Request.RemoteAddr != "" {
			ip = ctx.Request.RemoteAddr
		}

		if ip == "" {
			ip = ctx.Request.Header.Get("X-Real-IP")
		}
	}

	return strings.TrimSpace(ip)
}

// RenderStatus 仅向客户端输出状态码
//
// Deprecated: 直接使用 Error() 代替。
func (ctx *Context) RenderStatus(status int) {
	ctx.Error(status)
}
