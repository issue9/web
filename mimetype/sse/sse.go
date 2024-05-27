// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package sse [SSE] 的实现
//
// [SSE]: https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import "github.com/issue9/mux/v9/header"

// Mimetype sse 请求从服务端返回给客户端的内容类型
const Mimetype = header.EventStream
