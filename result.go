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

// Result 提供了一套用于描述向客户端反馈错误信息的机制。
//
// 对于错误代码的定义是根据 HTTP 状态码进行分类的，
// 比如所有与 400 有关的错误信息，都是以 400 * Scale 为基数的；
// 而与验证有关的都是以 401 * Scale 为基数的。Scale 为一个常量。
//
// 示例：
//  const(
//      BadRequest1 = http.StatusBadRequest * web.Scale + iota
//      BadRequest2
//      BadRequest3
//  )
//
//  func init(){
//      web.SetMessage(BadRequest1, "BadRequest1")
//      web.SetMessage(BadRequest2, "BadRequest2")
//      web.SetMessage(BadRequest3, "BadRequest3")
//  }
// 在 Result.IsError() 为 true 的情况下，也可以将其当作 error 使用。
type Result struct {
	Message string            `json:"message"`
	Code    int               `json:"code"`
	Detail  map[string]string `json:"detail,omitempty"`
}

// NewResult 声明一个新的 Result 实例
func NewResult(code int) *Result {
	return &Result{
		Code:    code,
		Message: Message(code),
		Detail:  make(map[string]string, 2),
	}
}

// NewResultWithDetail 声明一个带 Detail 内容的实例
func NewResultWithDetail(code int, detail map[string]string) *Result {
	return &Result{
		Code:    code,
		Message: Message(code),
		Detail:  detail,
	}
}

// Add 添加一条详细的错误信息。同名 field 会覆盖。
func (r *Result) Add(field, message string) *Result {
	r.Detail[field] = message
	return r
}

// Error error 接口
func (r *Result) Error() string {
	return r.Message
}

// HasDetail 是否包含详细的错误信息
func (r *Result) HasDetail() bool {
	return len(r.Detail) > 0
}

// IsError 当将 Result 当作 error 实例来用时，需要判断此值是否为 true。
func (r *Result) IsError() bool {
	return r.Status() >= 400
}

// Status 获取与其相对的 HTTP 状态码
func (r *Result) Status() int {
	return r.Code / Scale
}

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
