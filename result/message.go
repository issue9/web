// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"errors"
	"fmt"
)

// 保存所有的代码与消息对应关系
var messages = map[int]*message{}

// message 一个错误代码对应的内容
type message struct {
	message string // 消息信息
	status  int    // 对应的 HTTP 状态码
}

func getStatus(code int) int {
	for code > 999 {
		code /= 10
	}
	return code
}

// NewMessage 注册一条新的错误信息。
//
// 非协程安全，需要在程序初始化时添加所有的错误代码。
func NewMessage(code int, msg string) error {
	if code < 100 {
		return errors.New("ID 必须为大于等于 100 的值")
	}

	if len(msg) == 0 {
		return errors.New("参数 msg 不能为空值")
	}

	if _, found := messages[code]; found {
		return fmt.Errorf("重复的消息 ID: %v", code)
	}

	messages[code] = &message{message: msg, status: getStatus(code)}

	return nil
}

// NewMessages 注册错误代码。
// 非协程安全，需要在程序初始化时添加所有的错误代码。
func NewMessages(msgs map[int]string) error {
	for code, msg := range msgs {
		if err := NewMessage(code, msg); err != nil {
			return err
		}
	}

	return nil
}
