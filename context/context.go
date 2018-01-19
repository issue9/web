// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package context 对 HTTP 请求和输出作了简单的封装。
package context

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/issue9/web/encoding"
	"golang.org/x/text/transform"
)

var (
	// ErrUnsupportedContentType 表示用户提交的 Content-type
	// 报头中指定的编码或是字符集，不受当前系统的支持。
	ErrUnsupportedContentType = errors.New("不支持的 content-type 内容")

	// ErrClientNotAcceptable 表示客户端不接受当前系统指定的编码方式。
	ErrClientNotAcceptable = errors.New("客户端不接受当前的编码")
)

// Context 是对当前请求内容的封装，仅与当前请求相关。
type Context struct {
	w http.ResponseWriter
	r *http.Request

	// 输出内容到客户端时所使用的编码方式。
	marshal Marshal

	// 读取客户端内容时所使用的编码方式。这些编码方式必须是已经通过
	// AddUnmmarshal() 函数添加的。
	//
	// 此变量会通过 Content-Type 报头获取。
	unmarshal Unmarshal

	// 客户端内容的字符集，若为空，则表示为 utf-8
	//
	// 此值会通过 Content-Type 报头获取，
	// 且此字符集必须已经通过 AddCharset() 函数添加。
	inputCharset encoding.Charset

	// 输出到客户端的字符集，若为空，表示为 utf-8
	outputCharset encoding.Charset

	// 输出的编码方式
	outputEncodingName string

	// 输出的字符集
	outputCharsetName string

	// 从客户端获取的内容，已经解析为 utf-8 方式。
	body []byte
}

// New 声明一个 Context 实例。
//
// encodingName 指定出输出时的编码方式，此编码名称必须已经通过 AddMarshal 添加；
// charsetName 指定输出时的字符集，此字符集名称必须已经通过 AddCharset 添加；
// strict 若为 true，则会验证用户的 Accept 报头是否接受 encodingName 编码。
// 输入时的编码与字符集信息从报头 Content-Type 中获取，若未指定字符集，则默认为 utf-8
func New(w http.ResponseWriter, r *http.Request, encodingName, charsetName string, strict bool) (*Context, error) {
	marshal, found := marshals[encodingName]
	if !found {
		return nil, errors.New("encodingName 不存在")
	}

	outputCharset, found := charset[charsetName]
	if !found {
		return nil, errors.New("charsetName 不存在")
	}

	encName, charsetName := encoding.ParseContentType(r.Header.Get("Content-Type"))

	unmarshal, found := unmarshals[encName]
	if !found {
		return nil, ErrUnsupportedContentType
	}

	inputCharset, found := charset[charsetName]
	if !found {
		return nil, ErrUnsupportedContentType
	}

	if strict {
		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, encodingName) && !strings.Contains(accept, "*/*") {
			return nil, ErrClientNotAcceptable
		}
	}

	return &Context{
		w:                  w,
		r:                  r,
		marshal:            marshal,
		unmarshal:          unmarshal,
		inputCharset:       inputCharset,
		outputEncodingName: encodingName,
		outputCharset:      outputCharset,
		outputCharsetName:  charsetName,
	}, nil
}

// Body 获取用户提交的内容。
//
// 相对于 ctx.Request().Body，此函数可多次读取。
func (ctx *Context) Body() ([]byte, error) {
	if ctx.body == nil {
		bs, err := ioutil.ReadAll(ctx.r.Body)
		if err != nil {
			return nil, err
		}

		if ctx.inputCharset != nil {
			reader := transform.NewReader(bytes.NewReader(bs), ctx.inputCharset.NewDecoder())
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

	return ctx.unmarshal(body, v)
}

// Marshal 将 v 发送给客户端。
//
// NOTE: 若在 headers 中包含了 Content-Type，则会覆盖原来的 Content-Type 报头
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	ct := encoding.BuildContentType(ctx.outputEncodingName, ctx.outputCharsetName)
	if headers == nil {
		ctx.w.Header().Set("Content-Type", ct)
	} else if _, found := headers["Content-Type"]; !found {
		headers["Content-Type"] = ct

		for k, v := range headers {
			ctx.w.Header().Set(k, v)
		}
	}

	data, err := ctx.marshal(v)
	if err == nil {
		ctx.w.WriteHeader(status)

		if ctx.outputCharset != nil {
			w := transform.NewWriter(ctx.w, ctx.outputCharset.NewEncoder())
			_, err = w.Write(data)
		} else {
			_, err = ctx.w.Write(data)
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
	RenderStatus(ctx.w, status)
}

// Request 获取关联的 http.Request 实例
func (ctx *Context) Request() *http.Request {
	return ctx.r
}

// Response 获取关联的 http.ResponseWriter 实例
func (ctx *Context) Response() http.ResponseWriter {
	return ctx.w
}

// RenderStatus 仅向客户端输出状态码
func RenderStatus(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", encoding.BuildContentType(encoding.DefaultEncoding, encoding.DefaultCharset))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	fmt.Fprintln(w, http.StatusText(status))
}
