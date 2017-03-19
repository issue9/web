// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package types

import "net/http"

// Renderer 表示向客户端渲染的接口。
type Renderer interface {
	// Render 用于将 v 转换成相应编码的数据并写入到 w 中。
	//
	// code 为服务端返回的代码；
	// v 为需要输出的变量；
	// headers 用于指定额外的 Header 信息，若传递 nil，则表示没有。
	Render(code int, v interface{}, headers map[string]string)
}

// Reader 表示从客户端读取数据的接口。
type Reader interface {
	// 从客户端读取数据，若成功读取，则返回 true，
	// 否则返回 false，并向 w 输出相应的状态码和错误信息。
	Read(v interface{}) bool
}

// Context 对客户端输入输出的一个封装，此处仅是其接口的声明，
// 方便其它地方引用，直接的实现在 web 包中。
type Context interface {
	Reader
	Renderer

	// 返回 http.ResponseWriter 接口对象
	Response() http.ResponseWriter

	// 返回 *http.Request 对象
	Request() *http.Request
}
