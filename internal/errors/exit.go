// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errors

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/recovery"
)

// 表示一个 HTTP 状态码错误。
// panic 此类型的值，可以在 Revoery 中作特殊处理。
//
// 目前仅由 ExitCoroutine 使用，让框加以特定的状态码退出当前协程。
type httpStatus int

// 表示继续向上一级发出错误信息。
//
// 一般情况下表示要退出整个应用程序。
type exit struct {
	msg interface{}
}

// ExitCoroutine 以指定的状态码退出当前的协程
//
// status 表示输出的状态码，如果为 0，则不会作任何状态码输出。
func ExitCoroutine(status int) {
	panic(httpStatus(status))
}

// Recovery 生成一个 recovery.RecoverFunc 函数，用于抓获由 panic 触发的事件。
//
// debug 是否为调试模式，若是调试模式，则详细信息输出到客户端，否则输出到日志中。
func Recovery(debug bool) recovery.RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		// 通 httpStatus 退出的，并不能算是错误，所以此处不输出调用堆栈信息。
		if status, ok := msg.(httpStatus); ok {
			if status > 0 {
				render(w, int(status))
			}
			return
		}

		if obj, ok := msg.(exit); ok {
			panic(obj.msg)
		}

		render(w, http.StatusInternalServerError)
		if debug {
			w.Write([]byte(traceStack(3, msg)))
		} else {
			logs.Error(traceStack(3, msg))
		}
	}
}
