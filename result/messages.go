// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"fmt"

	xmessage "golang.org/x/text/message"
)

// GetResultFunc 用于生成 Result 接口对象的函数
type GetResultFunc func(status, code int, message string) Result

// Messages 保存所有的代码与消息对应关系
type Messages struct {
	get      func(int, int, string) Result
	messages map[int]*message
}

// NewMessages 声明 Messages 变量
func NewMessages(get GetResultFunc) *Messages {
	return &Messages{
		get:      get,
		messages: make(map[int]*message, 100),
	}
}

type message struct {
	message string // 消息信息
	status  int    // 对应的 HTTP 状态码
}

// New 查找指定代码的错误信息
func (m *Messages) New(code int) Result {
	msg, found := m.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return m.get(msg.status, code, msg.message)
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。如果为空，则表示返回原始值。
func (m *Messages) Messages(p *xmessage.Printer) map[int]string {
	msgs := make(map[int]string, len(m.messages))

	if p == nil {
		for code, msg := range m.messages {
			msgs[code] = msg.message
		}
	} else {
		for code, msg := range m.messages {
			msgs[code] = p.Sprintf(msg.message)
		}
	}

	return msgs
}

// 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码
// code 表示的是该错误的错误代码。
// msg 表示具体的错误描述内容。
func (m *Messages) newMessage(status, code int, msg string) {
	if len(msg) == 0 {
		panic("参数 msg 不能为空值")
	}

	if _, found := m.messages[code]; found {
		panic(fmt.Sprintf("重复的消息 ID: %d", code))
	}

	m.messages[code] = &message{message: msg, status: status}
}

// NewMessages 添加一组错误信息。
func (m *Messages) NewMessages(status int, msgs map[int]string) {
	for code, msg := range msgs {
		m.newMessage(status, code, msg)
	}
}
