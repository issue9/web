// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package errorhandler 对状态码错误的处理方式
package errorhandler

import (
	"net/http"

	"github.com/issue9/web/encoding"
)

var errorHandlers = map[int]func(http.ResponseWriter, int){}

// AddErrorHandler 添加针对特写状态码的错误处理函数
//
// 返回值表示是否添加成功
func AddErrorHandler(status int, f func(http.ResponseWriter, int)) bool {
	if _, found := errorHandlers[status]; found {
		return false
	}

	errorHandlers[status] = f
	return true
}

// SetErrorHandler 设置指定状态码对应的处理函数
//
// 有则修改，没有则添加
func SetErrorHandler(status int, f func(http.ResponseWriter, int)) {
	errorHandlers[status] = f
}

// defaultRender 用到的 content-type 类型
var errorContentType = encoding.BuildContentType("text/plain", "UTF-8")

// 仅向客户端输出状态码。
func defaultRender(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", errorContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	w.Write([]byte(http.StatusText(status) + "\n"))
}

// Render 向客户端输出指定状态码的错误内容。
func Render(w http.ResponseWriter, status int) {
	f, found := errorHandlers[status]
	if !found || f == nil {
		f = defaultRender
	}

	f(w, status)
}
