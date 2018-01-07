// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"

	"github.com/issue9/logs"
)

func (ctx *Context) logMessage(v []interface{}) string {
	w := new(bytes.Buffer)

	ws := func(val string) {
		_, err := w.WriteString(val)
		if err != nil {
			panic(err)
		}
	}

	if len(v) > 0 {
		if _, err := fmt.Fprint(w, v...); err != nil {
			panic(err)
		}
	}

	for i := 3; true; i++ {
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

// Critical 输出一条日志到 CRITICAL 日志通道，并向用户输出一个指定状态码的页面
// 若是输出日志的过程中出错，则 panic
func (ctx *Context) Critical(status int, v ...interface{}) {
	logs.CRITICAL().Output(2, ctx.logMessage(v))

	ctx.RenderStatus(status)
}

// Error 输出一条日志到 ERROR 日志通道，并向用户输出一个指定状态码的页面
// 若是输出日志的过程中出错，则 panic
func (ctx *Context) Error(status int, v ...interface{}) {
	logs.ERROR().Output(2, ctx.logMessage(v))

	ctx.RenderStatus(status)
}
