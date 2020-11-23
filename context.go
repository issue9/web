// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/issue9/middleware/v2/errorhandler"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/content"
)

// 需要作比较，所以得是经过 http.CanonicalHeaderKey 处理的标准名称。
var (
	contentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	contentLanguageKey = http.CanonicalHeaderKey("Content-Language")
)

// Context 是对当次 HTTP 请求内容的封装
type Context struct {
	server *Server

	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的媒体类型，以及名称
	OutputMimetype     content.MarshalFunc
	OutputMimetypeName string

	// 输出到客户端的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	OutputCharset     encoding.Encoding
	OutputCharsetName string

	// 客户端内容所使用的媒体类型
	InputMimetype content.UnmarshalFunc

	// 客户端内容所使用的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	InputCharset encoding.Encoding

	// 输出语言的相关设置项
	OutputTag     language.Tag
	LocalePrinter *message.Printer

	// 与当前对话相关的时区
	Location *time.Location

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式。
	//
	// 如果仅需要在多个请求中传递参数，可直接使用 Server.Vars。
	Vars map[interface{}]interface{}

	// 保存着从 http.Request.Body 中获取的内容。
	//
	// body 用于缓存从 http.Request.Body 中读取的内容；
	// read 表示是否需要从 http.Request.Body 读取内容。
	body []byte
	read bool
}

// NewContext 构建 *Context 实例
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return GetServer(r).NewContext(w, r)
}

// NewContext 构建 *Context 实例
//
// 如果 Accept 的内容与当前配置无法匹配，
// 则退出(panic)并输出 NotAcceptable 状态码。
func (srv *Server) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	checkError := func(name string, err error, status int) {
		if err == nil {
			return
		}

		srv.Logs().ERROR().Output(2, fmt.Sprintf("报头 %s 出错：%s\n", name, err.Error()))
		errorhandler.Exit(status)
	}

	header := r.Header.Get("Accept")
	outputMimetypeName, marshal, err := srv.mimetypes.Marshal(header)
	checkError("Accept", err, http.StatusNotAcceptable)

	header = r.Header.Get("Accept-Charset")
	outputCharsetName, outputCharset, err := content.AcceptCharset(header)
	checkError("Accept-Charset", err, http.StatusNotAcceptable)

	tag := content.AcceptLanguage(srv.catalog, r.Header.Get("Accept-Language"))

	ctx := &Context{
		server:             srv,
		Response:           w,
		Request:            r,
		OutputMimetype:     marshal,
		OutputMimetypeName: outputMimetypeName,
		OutputCharset:      outputCharset,
		OutputCharsetName:  outputCharsetName,
		OutputTag:          tag,
		LocalePrinter:      srv.NewLocalePrinter(tag),
		Location:           srv.Location(),
		Vars:               map[interface{}]interface{}{},
	}

	if header = r.Header.Get(contentTypeKey); header != "" {
		encName, charsetName, err := content.ParseContentType(header)
		checkError(contentTypeKey, err, http.StatusUnsupportedMediaType)

		ctx.InputMimetype, err = ctx.server.mimetypes.Unmarshal(encName)
		checkError(contentTypeKey, err, http.StatusUnsupportedMediaType)

		ctx.InputCharset, err = htmlindex.Get(charsetName)
		checkError(contentTypeKey, err, http.StatusUnsupportedMediaType)
	} else {
		ctx.read = true
	}

	return ctx
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request.Body，此函数可多次读取。
// 不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	if ctx.body, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		return nil, err
	}
	ctx.read = true

	if content.CharsetIsNop(ctx.InputCharset) {
		return ctx.body, nil
	}

	d := ctx.InputCharset.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(ctx.body), d)
	ctx.body, err = ioutil.ReadAll(reader)
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if ctx.InputMimetype != nil {
		return ctx.InputMimetype(body, v)
	}
	return nil
}

// Marshal 将 v 解码并发送给客户端
//
// 若 v 是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（JSON 中会输出 null），
// 可以使用 content.Nil 变量代替。
//
// NOTE: 如果需要指定一个特定的 Content-Type 和 Content-Language，
// 可以在 headers 中指定，否则使用当前的编码和语言名称。
//
// 通过 Marshal 输出的内容，即使 status 的值大于 399，
// 依然能正常输出 v 的内容，而不是转向 errorhandler 中的相关内容。
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	header := ctx.Response.Header()
	var contentTypeFound, contentLanguageFound bool
	for k, v := range headers {
		k = http.CanonicalHeaderKey(k)

		contentTypeFound = contentTypeFound || k == contentTypeKey
		contentLanguageFound = contentLanguageFound || k == contentLanguageKey
		header.Set(k, v)
	}

	if !contentTypeFound {
		ct := content.BuildContentType(ctx.OutputMimetypeName, ctx.OutputCharsetName)
		header.Set(contentTypeKey, ct)
	}

	if !contentLanguageFound && ctx.OutputTag != language.Und {
		header.Set(contentLanguageKey, ctx.OutputTag.String())
	}

	if v == nil {
		errorhandler.WriteHeader(ctx.Response, status)
		return nil
	}

	data, err := ctx.OutputMimetype(v)
	if err != nil {
		return err
	}

	// 注意 WriteHeader 调用顺序。
	// https://github.com/golang/go/issues/17083
	errorhandler.WriteHeader(ctx.Response, status)

	if content.CharsetIsNop(ctx.OutputCharset) {
		_, err = ctx.Response.Write(data)
		return err
	}

	w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
	if _, err = w.Write(data); err != nil {
		if err2 := w.Close(); err2 != nil {
			return fmt.Errorf("在处理错误 %w 时再次抛出错误 %s", err, err2.Error())
		}
		return err
	}
	return w.Close()
}

// Read 从客户端读取数据并转换成 v 对象
//
// 功能与 Unmarshal() 相同，只不过 Read() 在出错时，
// 会直接调用 Error() 处理：输出 422 的状态码，
// 并返回一个 false，告知用户转换失败。
// 如果是数据类型验证失败，则会输出以 code 作为错误代码的错误信息，
// 并返回 false，作为执行失败的通知。
func (ctx *Context) Read(v interface{}, code int) (ok bool) {
	if err := ctx.Unmarshal(v); err != nil {
		ctx.Error(http.StatusUnprocessableEntity, err)
		return false
	}

	if vv, ok := v.(Validator); ok {
		if errors := vv.CTXValidate(ctx); len(errors) > 0 {
			ctx.NewResultWithFields(code, errors).Render()
			return false
		}
	}

	return true
}

// Render 将 v 渲染给客户端
//
// 功能与 Marshal() 相同，只不过 Render() 在出错时，
// 会直接调用 Error() 处理，输出 500 的状态码。
//
// 如果需要具体控制出错后的处理方式，可以使用 Marshal 函数。
//
// 通过 Render 输出的内容，即使 status 的值大于 399，
// 依然能正常输出 v 的内容，而不是转向 errorhandler 中的相关内容，
// 但是渲染出错时，依然转换 errorhandler。
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	if err := ctx.Marshal(status, v, headers); err != nil {
		ctx.Error(http.StatusInternalServerError, err)
	}
}

// ClientIP 返回客户端的 IP 地址
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

// NotFound 404
//
// 接受统一的 errorhandler 模板支配
func (ctx *Context) NotFound() {
	ctx.Response.WriteHeader(http.StatusNotFound)
}

// NotImplemented 501
//
// 接受统一的 errorhandler 模板支配
func (ctx *Context) NotImplemented() {
	ctx.Response.WriteHeader(http.StatusNotImplemented)
}

// NewLocalePrinter 返回指定语言的 message.Printer
func (srv *Server) NewLocalePrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(srv.catalog))
}

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (ctx *Context) Now() time.Time {
	return time.Now().In(ctx.Location)
}

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location)
}
