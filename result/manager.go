// SPDX-License-Identifier: MIT

package result

import (
	"fmt"

	"golang.org/x/text/message"
)

// Manager 错误消息的管理
type Manager struct {
	messages map[int]*resultMessage
	builder  BuildFunc
}

type resultMessage struct {
	status int
	key    message.Reference
	values []interface{}
}

// NewManager 声明 Manager 实例
func NewManager(builder BuildFunc) *Manager {
	return &Manager{
		messages: make(map[int]*resultMessage, 20),
		builder:  builder,
	}
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。
func (mgr *Manager) Messages(p *message.Printer) map[int]string {
	msgs := make(map[int]string, len(mgr.messages))
	for code, msg := range mgr.messages {
		msgs[code] = p.Sprintf(msg.key, msg.values...)
	}
	return msgs
}

// AddMessage 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
func (mgr *Manager) AddMessage(status, code int, key message.Reference, v ...interface{}) {
	if _, found := mgr.messages[code]; found {
		panic(fmt.Sprintf("重复的消息 ID: %d", code))
	}
	mgr.messages[code] = &resultMessage{status: status, key: key, values: v}
}

// NewResult 返回 Result 实例
func (mgr *Manager) NewResult(p *message.Printer, code int) Result {
	msg, found := mgr.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return mgr.builder(msg.status, code, p.Sprintf(msg.key, msg.values...))
}

// NewResultWithFields 返回 Result 实例
func (mgr *Manager) NewResultWithFields(p *message.Printer, code int, fields Fields) Result {
	rslt := mgr.NewResult(p, code)

	for k, vals := range fields {
		rslt.Add(k, vals...)
	}

	return rslt
}
