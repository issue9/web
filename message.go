// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

// Scale 每个错误代码都是从 HTTP 状态码上放大此配数再进行累加的。
const Scale = 1000

// CodeNotExists 错误代码不存在时的提示信息
const CodeNotExists = "该错误代码不存在"

// 消息与代码的关联列表
var messages = make(map[int]string, 500)

// SetMessage 关联错误代码和错误信息。
func SetMessage(code int, message string) {
	messages[code] = message
}

// SetMessages 指量执行 SetMessage
func SetMessages(msgs map[int]string) {
	for code, msg := range msgs {
		SetMessage(code, msg)
	}
}

// Message 获取指定代码所表示的错误信息
func Message(code int) string {
	msg, found := messages[code]
	if found {
		return msg
	}

	return CodeNotExists
}
