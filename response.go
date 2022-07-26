// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/issue9/web/server/response"
)

// Status 仅向客户端输出状态码和报头
//
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
func Status(code int, kv ...string) Responser {
	return response.Status(code, kv...)
}

// Object 输出状态和对象至客户端
//
// body 表示需要输出的对象，该对象最终会被转换成相应的编码；
// kv 为报头，必须以偶数数量出现，奇数位为报头名，偶数位为对应的报头值；
func Object(status int, body any, kv ...string) Responser {
	return response.Object(status, body, kv...)
}

func Created(v any, location string) Responser {
	if location != "" {
		return Object(http.StatusCreated, v, "Location", location)
	}
	return Object(http.StatusCreated, v)
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
	return Status(status, "Retry-After", strconv.FormatUint(seconds, 10))
}

func RetryAt(status int, at time.Time) Responser {
	return Status(status, "Retry-After", at.UTC().Format(http.TimeFormat))
}

// Redirect 重定向至新的 URL
func Redirect(status int, url string) Responser {
	return Status(status, "Location", url)
}
