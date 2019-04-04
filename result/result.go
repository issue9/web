// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package result 提供了在非正常退出请求时的一些输出内容。
// 比如在给客户返回 400 时，可能会提供相应的错误字段内容等。
package result

// Result 提供了错误状态码的输出功能
type Result interface {
	error
	// 添加详细的内容
	Add(key, val string)

	// 设置详细内容
	Set(key, val string)

	// HTTP 状态码
	//
	// 最终会经此值作为 HTTP 状态会返回给用户
	Status() int
}
