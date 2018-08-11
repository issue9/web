// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/recovery"
	"github.com/issue9/utils"
)

// 表示一个 HTTP 状态码错误。
// panic 此类型的值，可以在 Revoery 中作特殊处理。
//
// 目前仅由 ExitCoroutine 使用，让框加以特定的状态码退出当前协程。
type httpStatus int

// Exit 以指定的状态码退出当前协程
//
// status 表示输出的状态码，如果为 0，则不会作任何状态码输出。
//
// Exit 最终是以 panic 的形式退出，所以如果你的代码里截获了 panic，
// 那么 Exit 并不能达到退出当前请求的操作。
//
// 与 Error 的不同在于：
// Error 不会主动退出当前协程，而 Exit 则会触发 panic，退出当前协程。
func Exit(status int) {
	panic(httpStatus(status))
}

// Recovery 生成一个 recovery.RecoverFunc 函数，用于捕获由 panic 触发的事件。
//
// debug 是否为调试模式，若是调试模式，则详细信息输出到客户端，否则输出到日志中。
func Recovery(debug bool) recovery.RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		// 通 httpStatus 退出的，并不能算是错误，所以此处不输出调用堆栈信息。
		if status, ok := msg.(httpStatus); ok {
			if status > 0 {
				Render(w, int(status))
			}
			return
		}

		Render(w, http.StatusInternalServerError)

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
		logs.Error(message)
	}
}
