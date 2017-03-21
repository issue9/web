// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrDuplicateMessageCode 表示消息 ID 有重复
var ErrDuplicateMessageCode = errors.New("重复的消息 ID")

// 保存所有的代码与消息对应关系
var messages = map[int]message{}

type message struct {
	message string // 消息信息
	status  int    // 对应的 HTTP 状态码
}

// 获取指定代码所表示的错误信息
func getMessage(code int) (message, error) {
	msg, found := messages[code]
	if found {
		return msg, nil
	}

	// 不存在相关的错误码
	return message{status: http.StatusInternalServerError, message: "未知错误"}, fmt.Errorf("错误代码[%v]不存在", code)
}

func getStatus(code int) int {
	for code > 999 {
		code /= 10
	}
	return code
}

// NewMessage 注册一条新的信息
func NewMessage(code int, msg string) error {
	if code < 100 {
		return fmt.Errorf("ID 必须为大于等于 %v 的值", 100)
	}

	if len(msg) == 0 {
		return errors.New("参数 msg 不能为空值")
	}

	if _, found := messages[code]; found {
		return ErrDuplicateMessageCode
	}

	messages[code] = message{message: msg, status: getStatus(code)}

	return nil
}

// NewMessages 批量注册信息
func NewMessages(msgs map[int]string) error {
	for code, msg := range msgs {
		if err := NewMessage(code, msg); err != nil {
			return err
		}
	}

	return nil
}
