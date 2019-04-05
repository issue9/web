// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"fmt"

	xmessage "golang.org/x/text/message"
)

// GetResultFunc 用于生成 Result 接口对象的函数
type GetResultFunc func(status, code int, message string) Result

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

type message struct {
	message string // 消息信息
	status  int    // 对应的 HTTP 状态码
}

// NewResult 查找指定代码的错误信息
func (app *App) NewResult(code int) Result {
	msg, found := app.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return app.getResult(msg.status, code, msg.message)
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。如果为空，则表示返回原始值。
func (app *App) Messages(p *xmessage.Printer) map[int]string {
	msgs := make(map[int]string, len(app.messages))

	if p == nil {
		for code, msg := range app.messages {
			msgs[code] = msg.message
		}
	} else {
		for code, msg := range app.messages {
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
func (app *App) newMessage(status, code int, msg string) {
	if len(msg) == 0 {
		panic("参数 msg 不能为空值")
	}

	if _, found := app.messages[code]; found {
		panic(fmt.Sprintf("重复的消息 ID: %d", code))
	}

	app.messages[code] = &message{message: msg, status: status}
}

// NewMessages 添加一组错误信息。
func (app *App) NewMessages(status int, msgs map[int]string) {
	for code, msg := range msgs {
		app.newMessage(status, code, msg)
	}
}
