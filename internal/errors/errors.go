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

	"github.com/issue9/web/encoding"
)

// RenderStatus 仅向客户端输出状态码。
// 编码和字符集均采用 encoding 的默认值。
//
// 一般情况下，用于输出错误的状态信息。
func RenderStatus(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", encoding.BuildContentType(encoding.DefaultMimeType, encoding.DefaultCharset))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	fmt.Fprintln(w, http.StatusText(status))
}

// TraceStack 返回调用者的堆栈信息
func TraceStack(level int, messages ...interface{}) string {
	var w strings.Builder

	ws := func(val string) {
		_, err := w.WriteString(val)
		if err != nil {
			panic(err)
		}
	}

	if len(messages) > 0 {
		if _, err := fmt.Fprint(&w, messages...); err != nil {
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
