// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"errors"
	"fmt"

	xmessage "golang.org/x/text/message"
)

// 保存所有的代码与消息对应关系
var messages = map[int]*message{}

// message 一个错误代码对应的内容
type message struct {
	message string // 消息信息
	status  int    // 对应的 HTTP 状态码
}

// Messages 错误信息列表
//
// 若需要特定语言的内容，可以调用 LocaleMessages() 函数获取。
func Messages() map[int]string {
	msgs := make(map[int]string, len(messages))
	for code, msg := range messages {
		msgs[code] = msg.message
	}

	return msgs
}

// LocaleMessages 本化地的错误信息列表
func LocaleMessages(p *xmessage.Printer) map[int]string {
	msgs := make(map[int]string, len(messages))
	for code, msg := range messages {
		msgs[code] = p.Sprintf(msg.message)
	}

	return msgs
}

func getStatus(code int) int {
	for code > 999 {
		code /= 10
	}
	return code
}

// NewMessage 注册一条新的错误信息。
//
// 功能与 NewStatusMessage 相同，但相较于 NewStatusMessage 函数，
// 少了一个表示 HTTP 状态码的 status 参数。
// NewMessage 中的状态码根据 code 的计算而来，将 code 一直被 10 相除，
// 真到值介于 100-999 之间，取该值作为 HTTP 状态码。
// 不判断该状态码是否真实存在于 RFC 定义中。
func NewMessage(code int, msg string) error {
	if code < 100 {
		return errors.New("ID 必须为大于等于 100 的值")
	}

	return NewStatusMessage(getStatus(code), code, msg)
}

// NewMessages 注册错误代码。
func NewMessages(msgs map[int]string) error {
	for code, msg := range msgs {
		if err := NewMessage(code, msg); err != nil {
			return err
		}
	}

	return nil
}

// NewStatusMessage 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码
// code 表示的是该错误的错误代码。
// msg 表示具体的错误描述内容。
func NewStatusMessage(status, code int, msg string) error {
	if len(msg) == 0 {
		return errors.New("参数 msg 不能为空值")
	}

	if _, found := messages[code]; found {
		return fmt.Errorf("重复的消息 ID: %d", code)
	}

	messages[code] = &message{message: msg, status: status}
	return nil
}

// NewStatusMessages 添加一组错误信息。
func NewStatusMessages(status int, msgs map[int]string) error {
	for code, msg := range msgs {
		if err := NewStatusMessage(status, code, msg); err != nil {
			return err
		}
	}

	return nil
}
