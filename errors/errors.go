// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errors

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
)

// HTTP 表示一个 HTTP 状态码错误
//
// 如果遇到不可处理的错误，可以 panic 此类型的值，其值为一个 HTTP 状态码，
// 则会立即以当前状态为返回结果，直接退出当前请求。
type HTTP int

func (err HTTP) Error() string {
	return http.StatusText(int(err))
}

// TraceStack 返回调用者的堆栈信息
func TraceStack(level int, messages ...interface{}) string {
	w := new(bytes.Buffer)

	ws := func(val string) {
		_, err := w.WriteString(val)
		if err != nil {
			panic(err)
		}
	}

	if len(messages) > 0 {
		if _, err := fmt.Fprint(w, messages...); err != nil {
			panic(err)
		}
	}

	for i := level; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		ws(file)
		ws(strconv.Itoa(line))
		ws("\n")
	}

	return w.String()
}
