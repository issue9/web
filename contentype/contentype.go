// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package contentype 包含了读取和输出特定格式数据的操作函数。
package contentype

import (
	"errors"
	"log"
	"net/http"

	"github.com/issue9/web/contentype/internal/json"
	"github.com/issue9/web/contentype/internal/xml"
)

// ErrUnsupported 表示不支持的编码类型
var ErrUnsupported = errors.New("不支持的编码类型")

// Renderer 表示向客户端渲染的接口。
type Renderer interface {
	// Render 用于将 v 转换成相应编码的数据并写入到 w 中。
	//
	// code 为服务端返回的代码。
	//
	// 若 v 的值是 string,[]byte，[]rune 则直接转换成字符串写入 w。为 nil 时，
	// 不输出任何内容，若需要输出一个空对象，请使用"{}"字符串；
	//
	// headers 用于指定额外的 Header 信息，若传递 nil，则表示没有。
	// NOTE: 会在返回的文件头信息中添加 Content-Type=application/json;charset=utf-8
	// 的信息，若想手动指定该内容，可通过在 headers 中传递同名变量来改变。
	Render(w http.ResponseWriter, r *http.Request, code int, v interface{}, headers map[string]string)
}

// Reader 表示从客户端读取数据的接口。
type Reader interface {
	// 从客户端读取数据，若成功读取，则返回 true，
	// 否则返回 false，并向 w 输出相应的状态码和错误信息。
	Read(w http.ResponseWriter, r *http.Request, v interface{}) bool
}

// ContentTyper 表示从客户端读取和输出时的相关编码解码操作。
type ContentTyper interface {
	Renderer
	Reader
}

// New 根据 encoding 类型，返回一个相应的 ContentType 接口实例。
//
// encoding 编码类型，支持 "json"、"xml" 两种类型，其它类型返回错误信息。
func New(encoding string, errlog *log.Logger) (ContentTyper, error) {
	switch encoding {
	case "json":
		return json.New(errlog), nil
	case "xml":
		return xml.New(errlog), nil
	default:
		return nil, ErrUnsupported
	}
}
