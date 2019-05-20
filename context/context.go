// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package context 用于处理单个请求的上下文关系。
package context

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/middleware/recovery/errorhandler"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/app"
	"github.com/issue9/web/mimetype"
)

// 需要作比较，所以得是经过 http.CanonicalHeaderKey 处理的标准名称。
var (
	contentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	contentLanguageKey = http.CanonicalHeaderKey("Content-Language")
)

// Context 是对当前请求内容的封装，仅与当前请求相关。
type Context struct {
	App *app.App

	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的媒体类型，以及名称
	OutputMimeType     mimetype.MarshalFunc
	OutputMimeTypeName string

	// 输出到客户端的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	OutputCharset     encoding.Encoding
	OutputCharsetName string

	// 客户端内容所使用的媒体类型。
	InputMimeType mimetype.UnmarshalFunc

	// 客户端内容所使用的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	InputCharset encoding.Encoding

	// 输出语言的相关设置项。
	OutputTag     language.Tag
	LocalePrinter *message.Printer

	// 保存着从 http.Request.Body 中获取的内容。
	//
	// body 用于缓存从 http.Request.Body 中读取的内容；
	// readed 表示是否需要从 http.Request.Body 读取内容。
	body   []byte
	readed bool
}

// New 根据当前请求内容生成 Context 对象
//
// 如果 Accept 的内容与当前配置无法匹配，
// 则退出(panic)并输出 NotAcceptable 状态码。
//
// NOTE: New 仅供框架内部使用，不保证兼容性。如果框架提供的 Context
// 不符合你的要求，那么请直接使用 &Context{} 指定相关的值构建对象。
func New(w http.ResponseWriter, r *http.Request, a *app.App) *Context {
	checkError := func(name string, err error, status int) {
		if err == nil {
			return
		}

		a.Logs().ERROR().Output(2, fmt.Sprintf("报头 %s 出错：%s\n", name, err.Error()))
		errorhandler.Exit(status)
	}

	header := r.Header.Get("Accept")
	outputMimeTypeName, marshal, err := a.Mimetypes().Marshal(header)
	checkError("Accept", err, http.StatusNotAcceptable)

	header = r.Header.Get("Accept-Charset")
	outputCharsetName, outputCharset, err := acceptCharset(header)
	checkError("Accept-Charset", err, http.StatusNotAcceptable)

	tag, err := acceptLanguage(r.Header.Get("Accept-Language"))
	checkError("Accept-Language", err, http.StatusNotAcceptable)

	ctx := &Context{
		App:                a,
		Response:           w,
		Request:            r,
		OutputMimeType:     marshal,
		OutputMimeTypeName: outputMimeTypeName,
		OutputCharset:      outputCharset,
		OutputCharsetName:  outputCharsetName,
		OutputTag:          tag,
		LocalePrinter:      a.LocalPrinter(tag),
	}

	if header = r.Header.Get(contentTypeKey); header != "" {
		encName, charsetName, err := parseContentType(header)
		checkError(contentTypeKey, err, http.StatusUnsupportedMediaType)

		ctx.InputMimeType, err = ctx.App.Mimetypes().Unmarshal(encName)
		checkError(contentTypeKey, err, http.StatusUnsupportedMediaType)

		ctx.InputCharset, err = htmlindex.Get(charsetName)
		checkError(contentTypeKey, err, http.StatusUnsupportedMediaType)
	} else {
		ctx.readed = true
	}

	return ctx
}

// Body 获取用户提交的内容。
//
// 相对于 ctx.Request.Body，此函数可多次读取。
// 不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.readed {
		return ctx.body, nil
	}

	if ctx.body, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		return nil, err
	}

	if charsetIsNop(ctx.InputCharset) {
		ctx.readed = true
		return ctx.body, nil
	}

	d := ctx.InputCharset.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(ctx.body), d)
	ctx.body, err = ioutil.ReadAll(reader)
	ctx.readed = true
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象。
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if ctx.InputMimeType != nil {
		return ctx.InputMimeType(body, v)
	}

	return nil
}

// Marshal 将 v 解码并发送给客户端。
//
// 若 v 是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（JSON 中会输出 null），
// 可以使用 mimetype.Nil 变量代替。
//
// NOTE: 如果需要指定一个特定的 Content-Type 和 Content-Language，
// 可以在 headers 中指定，否则使用当前的编码和语言名称。
//
// 通过 Render 输出的内容，即使 status 的值大于 399，
// 依赖能正常输出 v 的内容，而不是转向 errorhandler 中的相关内容。
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	header := ctx.Response.Header()
	var contentTypeFound, contentLanguageFound bool
	for k, v := range headers {
		k = http.CanonicalHeaderKey(k)

		contentTypeFound = (contentTypeFound || k == contentTypeKey)
		contentLanguageFound = (contentLanguageFound || k == contentLanguageKey)
		header.Set(k, v)
	}

	if !contentTypeFound {
		ct := buildContentType(ctx.OutputMimeTypeName, ctx.OutputCharsetName)
		header.Set(contentTypeKey, ct)
	}

	if !contentLanguageFound && ctx.OutputTag != language.Und {
		header.Set(contentLanguageKey, ctx.OutputTag.String())
	}

	if v == nil {
		errorhandler.WriteHeader(ctx.Response, status)
		return nil
	}

	data, err := ctx.OutputMimeType(v)
	if err != nil {
		return err
	}

	// 注意 WriteHeader 调用顺序。
	// https://github.com/golang/go/issues/17083
	errorhandler.WriteHeader(ctx.Response, status)

	if charsetIsNop(ctx.OutputCharset) {
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
//
// 通过 Render 输出的内容，即使 status 的值大于 399，
// 依赖能正常输出 v 的内容，而不是转向 errorhandler 中的相关内容。
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
	if ip == "" && ctx.Request.RemoteAddr != "" {
		ip = ctx.Request.RemoteAddr
	}
	if ip == "" {
		ip = ctx.Request.Header.Get("X-Real-IP")
	}

	return strings.TrimSpace(ip)
}

// Created 201
func (ctx *Context) Created(v interface{}, location string) {
	if location == "" {
		ctx.Render(http.StatusCreated, v, nil)
	} else {
		ctx.Render(http.StatusCreated, v, map[string]string{
			"Location": location,
		})
	}
}

// NoContent 204
func (ctx *Context) NoContent() {
	errorhandler.Exit(http.StatusNoContent)
}

// ResetContent 205
func (ctx *Context) ResetContent() {
	errorhandler.Exit(http.StatusResetContent)
}

// NotImplemented 501
func (ctx *Context) NotImplemented() {
	// 接收统一的 errorhandlers 模板支配
	ctx.Response.WriteHeader(http.StatusNotImplemented)
}
