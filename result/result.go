// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package result 管理错误状态码与其对应的消息内容
package result

// Result 输出错误结果的接口
type Result interface {
	// 添加详细的内容
	Add(key, val string)

	Set(key, val string)

	// HTTP 状态码
	Status() int
}
