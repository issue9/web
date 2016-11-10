// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

// 每个错误代码都是从 HTTP 状态码上放大此配数再进行累加的。
const scale = 1000

const codeNotExists = "该错误代码不存"

var (
	messages = map[int]string{}
	indexes  = map[int]int{}
)

// RegisterMessage 在某一个 HTTP 状态码下注册一个新的错误信息并返回表示该信息的代码。
//
// NOTE: 不应该在多个协程中调用 RegisterMessage，以免每次重启程序，分配的代码都不一样。
func RegisterMessage(status int, message string) int {
	index, found := indexes[status]
	if !found {
		index = status * scale
	} else {
		index++
	}

	messages[index] = message
	indexes[status] = index
	return index
}

// Message 获取指定代码所表示的错误信息
func Message(code int) string {
	msg, found := messages[code]
	if found {
		return msg
	}

	return codeNotExists
}
