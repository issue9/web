// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package context 对 HTTP 请求和输出作了简单的封装。
package context

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/issue9/web/encoding"
	"golang.org/x/text/transform"
)

// Context 是对当前请求内容的封装，仅与当前请求相关。
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的编码方式，以及名称
	OutputEncoding     encoding.Marshal
	OutputEncodingName string

	// 输出到客户端的字符集，若为空，表示为 utf-8
	OutputCharset     encoding.Charset
	OutputCharsetName string

	// InputEncoding 读取客户端内容时所使用的编码方式。
	InputEncoding encoding.Unmarshal

	// 客户端内容的字符集，若为空，则表示为 utf-8
	//
	// 此值会通过 Content-Type 报头获取，
	// 且此字符集必须已经通过 AddCharset() 函数添加。
	InputCharset encoding.Charset

	// 从客户端获取的内容，已经解析为 utf-8 方式。
	body []byte
}

// Body 获取用户提交的内容。
//
// 相对于 ctx.Request().Body，此函数可多次读取。
func (ctx *Context) Body() ([]byte, error) {
	if ctx.body == nil {
		bs, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			return nil, err
		}

		if ctx.InputCharset != nil {
			reader := transform.NewReader(bytes.NewReader(bs), ctx.InputCharset.NewDecoder())
			bs, err = ioutil.ReadAll(reader)
			if err != nil {
				return nil, err
			}
		}

		ctx.body = bs
	}

	return ctx.body, nil
}

// Unmarshal 将提交的内容转换成 v 对象。
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	return ctx.InputEncoding(body, v)
}

// Marshal 将 v 发送给客户端。
//
// NOTE: 若在 headers 中包含了 Content-Type，则会覆盖原来的 Content-Type 报头
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	ct := encoding.BuildContentType(ctx.OutputEncodingName, ctx.OutputCharsetName)
	if headers == nil {
		ctx.Response.Header().Set("Content-Type", ct)
	} else if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = ct

		for k, v := range headers {
			ctx.Response.Header().Set(k, v)
		}
	}

	data, err := ctx.OutputEncoding(v)
	if err == nil {
		ctx.Response.WriteHeader(status)

		if ctx.OutputCharset != nil {
			w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
			_, err = w.Write(data)
		} else {
			_, err = ctx.Response.Write(data)
		}
	}

	return err
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
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	if err := ctx.Marshal(status, v, headers); err != nil {
		ctx.Error(http.StatusInternalServerError, err)
	}
}

// RenderStatus 仅向客户端输出状态码
func (ctx *Context) RenderStatus(status int) {
	RenderStatus(ctx.Response, status)
}

// RenderStatus 仅向客户端输出状态码
func RenderStatus(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", encoding.BuildContentType(encoding.DefaultEncoding, encoding.DefaultCharset))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	fmt.Fprintln(w, http.StatusText(status))
}
