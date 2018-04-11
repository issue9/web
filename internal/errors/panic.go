// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errors

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/recovery"
)

// 表示一个 HTTP 状态码错误
//
// 如果遇到不可处理的错误，可以 panic 此类型的值，其值为一个 HTTP 状态码，
// 则会立即以当前状态为返回结果，直接退出当前请求。
type httpStatus int

// Panic 以指定的状态码抛出异常
func Panic(status int) {
	panic(httpStatus(status))
}

// Recovery 生成一个 recovery.RecoverFunc 函数
// 用于抓获由 Panic() 触发的事件。
//
// debug 是否为调试模式，或是调试模式，则详细信息输出到客户端，否则输出到日志中。
func Recovery(debug bool) recovery.RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		if err, ok := msg.(httpStatus); ok {
			renderStatus(w, int(err))
		} else {
			renderStatus(w, http.StatusInternalServerError)
		}

		if debug {
			w.Write([]byte(traceStack(3, msg)))
		} else {
			logs.Error(msg)
		}
	}
}
