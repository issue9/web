// SPDX-License-Identifier: MIT

// Package result 统一了在出错时返回给用户的错误信息处理方式
package result

import (
	"fmt"

	xmessage "golang.org/x/text/message"
)

// BuildResultFunc 用于生成 Result 接口对象的函数
type BuildResultFunc func(status, code int, message string) Result

// Result 自定义错误代码的实现接口
//
// 用户可以根据自己的需求，在出错时，展示自定义的错误码以及相关的错误信息格式。
// 只要该对象实现了 Result 接口即可。
//
// 比如类似以下的错误内容：
//  {
//      'message': 'error message',
//      'code': 4000001,
//      'detail':[
//          {'field': 'username': 'message': '已经存在相同用户名'},
//          {'field': 'username': 'message': '已经存在相同用户名'},
//      ]
//  }
//
// 可以在 default.go 查看 Result 的实现方式。
type Result interface {
	// 添加详细的内容
	Add(key, val string)

	// 设置详细的内容
	Set(key, val string)

	// 是否存在详细的错误信息
	//
	// 如果有通过 Add 添加内容，那么应该返回 true
	HasDetail() bool

	// HTTP 状态码
	//
	// 最终会经此值作为 HTTP 状态会返回给用户
	Status() int
}

type message struct {
	message string
	status  int // 对应的 HTTP 状态码
}

// Results 管理 Result 的集合
type Results struct {
	messages map[int]*message
	build    BuildResultFunc
}

// NewResults 声明 *Results 实例
func NewResults(b BuildResultFunc) *Results {
	return &Results{
		messages: make(map[int]*message, 100),
		build:    b,
	}
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。如果为空，则表示返回原始值。
func (rslt *Results) Messages(p *xmessage.Printer) map[int]string {
	msgs := make(map[int]string, len(rslt.messages))

	if p == nil {
		for code, msg := range rslt.messages {
			msgs[code] = msg.message
		}
	} else {
		for code, msg := range rslt.messages {
			msgs[code] = p.Sprintf(msg.message)
		}
	}

	return msgs
}

// AddMessages 添加一组错误信息。
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
// msgs 中，键名表示的是该错误的错误代码；
// 键值表示具体的错误描述内容。
func (rslt *Results) AddMessages(status int, msgs map[int]string) {
	for code, msg := range msgs {
		if msg == "" {
			panic("参数 msg 不能为空值")
		}

		if _, found := rslt.messages[code]; found {
			panic(fmt.Sprintf("重复的消息 ID: %d", code))
		}

		rslt.messages[code] = &message{message: msg, status: status}
	}
}

// NewResult 查找指定代码的错误信息
func (rslt *Results) NewResult(code int) Result {
	msg, found := rslt.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return rslt.build(msg.status, code, msg.message)
}
