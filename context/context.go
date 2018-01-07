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

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
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
	inputCharset encoding.Encoding

	// 输出到客户端的字符集，若为空，表示为 utf-8
	outputCharset encoding.Encoding

	// 输出的编码方式
	outputEncoding string

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
	m, found := marshals[encodingName]
	if !found {
		return nil, errors.New("encodingName 不存在")
	}

	c, found := charset[charsetName]
	if !found {
		return nil, errors.New("charsetName 不存在")
	}

	ctx := &Context{
		w:                 w,
		r:                 r,
		outputEncoding:    encodingName,
		marshal:           m,
		outputCharset:     c,
		outputCharsetName: charsetName,
	}

	if r.Method != http.MethodGet {
		encName, charsetName := parseContentType(r.Header.Get("Content-Type"))

		enc, found := unmarshals[encName]
		if !found {
			return nil, errors.New("不支持的媒体类型")
		}
		ctx.unmarshal = enc

		c, found := charset[charsetName]
		if !found {
			return nil, errors.New("不支持的字符集")
		}
		ctx.inputCharset = c
	}

	if strict {
		accept := r.Header.Get("Accept")
		if strings.Index(accept, encodingName) < 0 && strings.Index(accept, "*/*") < 0 {
			return nil, errors.New("客户端不支持当前的编码方式")
		}
	}

	return ctx, nil
}

// Body 读取所有内容，相对于 ctx.Request().Body，此值可多次读取
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

			ctx.body = bs
		}
	}

	return ctx.body, nil
}

// Read 从客户端读取数据并转换成 v 对象
func (ctx *Context) Read(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	return ctx.unmarshal(body, v)
}

// Render 将 v 渲染给客户端。
// 若是出错，则会尝试调用 ctx.Error() 输出错误信息。
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	ct := buildContentType(ctx.outputEncoding, ctx.outputCharsetName)
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

	if err != nil {
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
	w.Header().Set("Content-Type", buildContentType(DefaultEncoding, DefaultCharset))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	fmt.Fprintln(w, http.StatusText(status))
}
