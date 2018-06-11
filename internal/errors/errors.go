// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package errors 包含了一些公用的错误处理函数和类型定义
package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/web/encoding"
)

// 错误状态下，输出的 content-type 报头内容
var errorContentType = encoding.BuildContentType("text/plain", encoding.DefaultCharset)

// Error 输出一条日志到 ERROR 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func Error(level int, w http.ResponseWriter, status int, v ...interface{}) {
	if len(v) > 0 {
		logs.ERROR().Output(level, traceStack(level, v...))
	}

	render(w, status)
}

// Critical 输出一条日志到 CRITICAL 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func Critical(level int, w http.ResponseWriter, status int, v ...interface{}) {
	if len(v) > 0 {
		logs.CRITICAL().Output(level, traceStack(level, v...))
	}

	render(w, status)
}

// Errorf 输出一条日志到 ERROR 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func Errorf(level int, w http.ResponseWriter, status int, format string, v ...interface{}) {
	if len(v) > 0 {
		logs.ERROR().Output(level, traceStack(level, fmt.Sprintf(format, v...)))
	}

	render(w, status)
}

// Criticalf 输出一条日志到 CRITICAL 日志通道，并向用户输出一个指定状态码的页面。
//
// 若是输出日志的过程中出错，则 panic
// 若没有错误信息，则仅向客户端输出一条状态码信息。
func Criticalf(level int, w http.ResponseWriter, status int, format string, v ...interface{}) {
	if len(v) > 0 {
		logs.CRITICAL().Output(level, traceStack(level, fmt.Sprintf(format, v...)))
	}

	render(w, status)
}

// 仅向客户端输出状态码。
// 编码和字符集均采用 encoding 的默认值。
func render(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", errorContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	w.Write([]byte(http.StatusText(status) + "\n"))
}

// 返回调用者的堆栈信息
func traceStack(level int, messages ...interface{}) string {
	var w strings.Builder

	ws := func(val string) {
		_, err := w.WriteString(val)
		if err != nil {
			// BUG(caixw) 此处的 panic 会被 Recovery 接收，而 Recovery
			// 又会再次调用 traceStack，所以如果此处的 panic 成功触发，
			// 而必然造成死循环。
			panic(err)
		}
	}

	if len(messages) > 0 {
		if _, err := fmt.Fprintln(&w, messages...); err != nil {
			panic(err)
		}
	}

	for i := level; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		ws(file)
		ws(":")
		ws(strconv.Itoa(line))
		ws("\n")
	}

	return w.String()
}
