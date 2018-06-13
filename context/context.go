// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package context 用于处理单个请求的上下文关系。
package context

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/logs"
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

// New 根据当前请求内容生成 Context 对象
//
// 如果 Accept 的内容与当前配置无法匹配，
// 则退出(panic)并输出 NotAcceptable 状态码。
//
// 一些特殊类型的请求，比如上传操作等，可能无法直接通过 New 构造一个合适的 Context，
// 此时可以直接使用 &Context{} 的方法手动指定 Context 的各个变量值。
func New(w http.ResponseWriter, r *http.Request) *Context {
	unmarshal, charset, err := encoding.ContentType(r.Header.Get("Content-Type"))
	if err != nil {
		logs.Error(err)
		Exit(http.StatusUnsupportedMediaType)
	}

	outputMimeType, marshal, err := encoding.AcceptMimeType(r.Header.Get("Accept"))
	if err != nil {
		logs.Error(err)
		Exit(http.StatusNotAcceptable)
	}

	outputCharsetName, outputCharset, err := encoding.AcceptCharset(r.Header.Get("Accept-Charset"))
	if err != nil {
		logs.Error(err)
		Exit(http.StatusNotAcceptable)
	}

	return &Context{
		Response:           w,
		Request:            r,
		OutputMimeType:     marshal,
		OutputMimeTypeName: outputMimeType,
		InputMimeType:      unmarshal,
		InputCharset:       charset,
		OutputCharset:      outputCharset,
		OutputCharsetName:  outputCharsetName,
	}
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

	if ctx.InputCharset == nil || ctx.InputCharset == xencoding.Nop {
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

var contentTypeKey = http.CanonicalHeaderKey("Content-type")

// Marshal 将 v 解码并发送给客户端。
//
// 若 v 是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（json 中会输出 null），可以使用 Nil 变量代替。
//
// NOTE: 如果需要指定一个特定的 Content-Type，可以在 headers 中指定，
// 否则使用当前的编码名称。
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	found := false
	for k, v := range headers {
		// strings.ToLower 的性能未必有 http.CanonicalHeaderKey 好，
		// 所以直接使用了 http.CanonicalHeaderKey 作转换。
		if http.CanonicalHeaderKey(k) == contentTypeKey {
			found = true
		}
		ctx.Response.Header().Set(k, v)

	}
	if !found {
		ct := encoding.BuildContentType(ctx.OutputMimeTypeName, ctx.OutputCharsetName)
		ctx.Response.Header().Set("Content-Type", ct)
	}

	if v == nil {
		ctx.Response.WriteHeader(status)
		return nil
	}

	data, err := ctx.OutputMimeType(v)
	if err != nil {
		return err
	}

	ctx.Response.WriteHeader(status)

	if ctx.OutputCharset == nil || ctx.OutputCharset == xencoding.Nop {
		_, err = ctx.Response.Write(data)
		return err
	}

	w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
	if _, err = w.Write(data); err != nil {
		w.Close()
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
