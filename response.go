// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/issue9/web/server"
)

var objectPool = &sync.Pool{New: func() any { return &object{} }}

type (
	Responser = server.Responser

	object struct {
		status  int
		body    any
		headers map[string]string
	}
)

func Status(status int) Responser { return server.Status(status) }

func Object(status int, body interface{}, headers map[string]string) Responser {
	o := objectPool.Get().(*object)
	o.status = status
	o.body = body
	o.headers = headers
	return o
}

func (o *object) Apply(ctx *Context) {
	for k, v := range o.headers {
		ctx.Header().Set(k, v)
	}

	if err := ctx.Marshal(o.status, o.body); err != nil {
		ctx.Logs().ERROR().Error(err)
	}
	objectPool.Put(o)
}

func Created(v any, location string) Responser {
	if location != "" {
		return Object(http.StatusCreated, v, map[string]string{"Location": location})
	}
	return Object(http.StatusCreated, v, nil)
}

// OK 返回 200 状态码下的对象
func OK(v any) Responser { return Object(http.StatusOK, v, nil) }

func NotFound() Responser { return Status(http.StatusNotFound) }

func NoContent() Responser { return Status(http.StatusNoContent) }

func NotImplemented() Responser { return Status(http.StatusNotImplemented) }

// RetryAfter 返回 Retry-After 报头内容
//
// 一般适用于 301 和 503 报文。
//
// status 表示返回的状态码；seconds 表示秒数，如果想定义为时间格式，
// 可以采用 RetryAt 函数，两个功能是相同的，仅是时间格式上有差别。
func RetryAfter(status int, seconds uint64) Responser {
	return Object(status, nil, map[string]string{
		"Retry-After": strconv.FormatUint(seconds, 10),
	})
}

func RetryAt(status int, at time.Time) Responser {
	return Object(status, nil, map[string]string{
		"Retry-After": at.UTC().Format(http.TimeFormat),
	})
}

// Redirect 重定向至新的 URL
func Redirect(status int, url string) Responser {
	return Object(status, nil, map[string]string{"Location": url})
}
