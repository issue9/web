// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package errors 对状态码错误的处理方式
package errors

import (
	"fmt"
	"net/http"

	"github.com/issue9/web/encoding"
)

var errorHandlers = map[int]func(http.ResponseWriter, int){}

// AddErrorHandler 添加针对特写状态码的错误处理函数
func AddErrorHandler(f func(http.ResponseWriter, int), status ...int) error {
	for _, s := range status {
		if _, found := errorHandlers[s]; found {
			return fmt.Errorf("状态码 %d 已经存在", s)
		}

		errorHandlers[s] = f
	}

	return nil
}

// SetErrorHandler 设置指定状态码对应的处理函数
//
// 有则修改，没有则添加
func SetErrorHandler(f func(http.ResponseWriter, int), status ...int) {
	for _, s := range status {
		errorHandlers[s] = f
	}
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
	if !found {
		if f, found = errorHandlers[0]; !found || f == nil {
			f = defaultRender
		}
	} else if f == nil {
		f = defaultRender
	}

	f(w, status)
}
