// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/issue9/web/server"
)

type Responser = server.Responser

func Status(status int) Responser { return server.Status(status) }

func Object(status int, body interface{}) *server.Object { return server.Body(status, body) }

func Created(v any, location string) Responser {
	resp := Object(http.StatusCreated, v)
	if location != "" {
		resp.Header("Location", location)
	}
	return resp
}

// OK 返回 200 状态码下的对象
func OK(v any) Responser { return Object(http.StatusOK, v) }

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
	return Object(status, nil).Header("Retry-After", strconv.FormatUint(seconds, 10))
}

func RetryAt(status int, at time.Time) Responser {
	return Object(status, nil).Header("Retry-After", at.UTC().Format(http.TimeFormat))
}
