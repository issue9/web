// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package contentype 包含了读取和输出特定格式数据的操作函数。
//
// 原则上，所有从客户端读取数据或是向客户端输出数据，都应该是
// 通过此包的相关功能函数实现的。
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
	// code 为服务端返回的代码；
	// v 为需要输出的变量；
	// headers 用于指定额外的 Header 信息，若传递 nil，则表示没有。
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
