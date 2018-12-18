// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"net/http"

	"github.com/issue9/middleware/recovery"
	"github.com/issue9/utils"

	"github.com/issue9/web/internal/exit"
)

// ErrorHandler 错误处理函数
type ErrorHandler func(http.ResponseWriter, int)

// AddErrorHandler 添加针对特写状态码的错误处理函数
func (app *App) AddErrorHandler(f ErrorHandler, status ...int) error {
	for _, s := range status {
		if _, found := app.errorHandlers[s]; found {
			return fmt.Errorf("状态码 %d 已经存在", s)
		}

		app.errorHandlers[s] = f
	}

	return nil
}

// SetErrorHandler 设置指定状态码对应的处理函数
//
// 有则修改，没有则添加
func (app *App) SetErrorHandler(f ErrorHandler, status ...int) {
	for _, s := range status {
		app.errorHandlers[s] = f
	}
}

// 仅向客户端输出状态码。
func defaultRender(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// RenderError 向客户端输出指定状态码的错误内容。
func (app *App) RenderError(w http.ResponseWriter, status int) {
	f, found := app.errorHandlers[status]
	if !found {
		if f, found = app.errorHandlers[0]; !found || f == nil {
			f = defaultRender
		}
	} else if f == nil {
		f = defaultRender
	}

	f(w, status)
}

// ExitContext 退出当前的请求处理协程
func ExitContext(status int) {
	exit.Context(status)
}

// 生成一个 recovery.RecoverFunc 函数，用于捕获由 panic 触发的事件。
//
// debug 是否为调试模式，若是调试模式，则详细信息输出到客户端，否则输出到日志中。
func (app *App) recovery(debug bool) recovery.RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		// 通 httpStatus 退出的，并不能算是错误，所以此处不输出调用堆栈信息。
		if status, ok := msg.(exit.HTTPStatus); ok {
			if status > 0 {
				app.RenderError(w, int(status))
			}
			return
		}

		app.RenderError(w, http.StatusInternalServerError)

		message, err := utils.TraceStack(3, msg)
		if err != nil {
			panic(err)
		}

		if debug {
			_, err := w.Write([]byte(message))
			if err != nil { // 输出错误时再次出错，则 panic，退出整个程序
				panic(err)
			}
		}
		app.Error(message)
	}
}
