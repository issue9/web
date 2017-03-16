// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package content 包含了对 HTTP 内容在各种编码下的读写操作。
package content

import (
	"errors"
	"net/http"
)

// Content 包含了对 HTTP 内容的读写操作。
type Content interface {
	// Render 向客户端渲染的函数声明。
	//
	// code 为服务端返回的代码；
	// v 为需要输出的变量；
	// headers 用于指定额外的 Header 信息，若传递 nil，则表示没有。
	//
	// NOTE: Render 最终会被 Context.Render 引用，
	// 所以在 Render 的实现者中不能调用 Context.Render 函数
	Render(w http.ResponseWriter, r *http.Request, code int, v interface{}, headers map[string]string)

	// Read 从客户端读取数据的函数声明。
	//
	// 若成功读取，则返回 true，否则返回 false，并向客户端输出相应的状态码和错误信息。
	//
	// NOTE: Read 最终会被 Context.Read 引用，
	// 所以在 Read 的实现者中不能调用 Context.Read 函数
	Read(w http.ResponseWriter, r *http.Request, v interface{}) bool
}

// New 声明一个 Content 实例。
//
// conf 对 envelope 功能的配置，若不需要直接直接传递 nil 值。
func New(conf *Config) (Content, error) {
	if conf == nil {
		return nil, errors.New("参数 conf 不能为空")
	}

	if err := conf.Check(); err != nil {
		return nil, err
	}

	switch conf.ContentType {
	case "json":
		return newJSON(conf), nil
	case "xml":
		return newXML(conf), nil
	default:
		return nil, errors.New("不支持的 contenttype")
	}
}
