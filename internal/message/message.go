// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package message

import (
	"errors"
	"fmt"
	"net/http"
)

// 保存所有的代码与消息对应关系
var messages = map[int]*Message{}

// Message 一个错误代码对应的内容
type Message struct {
	Message string // 消息信息
	Status  int    // 对应的 HTTP 状态码
}

// GetMessage 获取指定代码所表示的错误信息
func GetMessage(code int) (*Message, error) {
	msg, found := messages[code]
	if found {
		return msg, nil
	}

	// 不存在相关的错误码，返回 500 错误
	return &Message{Status: http.StatusInternalServerError, Message: "未知错误"}, fmt.Errorf("错误代码[%v]不存在", code)
}

func getStatus(code int) int {
	for code > 999 {
		code /= 10
	}
	return code
}

// Register 注册一条新的错误信息。
// 非协程安全，需要在程序初始化时添加所有的错误代码。
//
// code 必须为一个大于 100 的整数。
func Register(code int, msg string) error {
	if code < 100 {
		return fmt.Errorf("ID 必须为大于等于 %v 的值", 100)
	}

	if len(msg) == 0 {
		return errors.New("参数 msg 不能为空值")
	}

	if _, found := messages[code]; found {
		return fmt.Errorf("重复的消息 ID: %v", code)
	}

	messages[code] = &Message{Message: msg, Status: getStatus(code)}

	return nil
}

// Registers 批量注册信息
func Registers(msgs map[int]string) error {
	for code, msg := range msgs {
		if err := Register(code, msg); err != nil {
			return err
		}
	}

	return nil
}

// Clean 清除内容
func Clean() {
	messages = map[int]*Message{}
}
