// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package messages 管理错误状态码与其对应的消息内容
package messages

import (
	"fmt"

	xmessage "golang.org/x/text/message"
)

// Messages 保存所有的代码与消息对应关系
type Messages struct {
	messages map[int]*Message
}

// New 声明 Messages 变量
func New() *Messages {
	return &Messages{
		messages: make(map[int]*Message, 100),
	}
}

// Message 一个错误代码对应的内容
type Message struct {
	Message string // 消息信息
	Status  int    // 对应的 HTTP 状态码
}

// Message 查找指定代码的错误信息
func (m *Messages) Message(code int) (*Message, bool) {
	msg, found := m.messages[code]
	return msg, found
}

// Messages 错误信息列表
//
// 若需要特定语言的内容，可以调用 LocaleMessages() 函数获取。
func (m *Messages) Messages() map[int]string {
	msgs := make(map[int]string, len(m.messages))
	for code, msg := range m.messages {
		msgs[code] = msg.Message
	}

	return msgs
}

// LocaleMessages 本化地的错误信息列表
func (m *Messages) LocaleMessages(p *xmessage.Printer) map[int]string {
	msgs := make(map[int]string, len(m.messages))
	for code, msg := range m.messages {
		msgs[code] = p.Sprintf(msg.Message)
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

	m.messages[code] = &Message{Message: msg, Status: status}
}

// NewMessages 添加一组错误信息。
func (m *Messages) NewMessages(status int, msgs map[int]string) {
	for code, msg := range msgs {
		m.newMessage(status, code, msg)
	}
}
