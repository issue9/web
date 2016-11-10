// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// result 提供了一套用于描述向客户端反馈错误信息的机制。
//
// 对于错误代码的定义是根据 HTTP 状态码进行分类的，
// 比如所有与 400 有关的错误信息，都是以 400 * Scale 为基数的；
// 而与验证有关的都是以 401 * Scale 为基数的。Scale 为一个常量。
//
// 示例：
//  const(
//      BadRequest1 = http.StatusBadRequest * result.Scale + iota
//      BadRequest2
//      BadRequest3
//  )
//
//  func init(){
//      result.SetMessage(BadRequest1, "BadRequest1")
//      result.SetMessage(BadRequest2, "BadRequest2")
//      result.SetMessage(BadRequest3, "BadRequest3")
//  }
package result

// Result 表示在最终向用户展示的提示信息。同时也实现了 error 接口，
// 在 Result.IsError() 返回 true 的情况下，其实例可当作 error。
type Result struct {
	Message string            `json:"message"`
	Code    int               `json:"code"`
	Detail  map[string]string `json:"detail,omitempty"`
}

// New 声明一个新的 Result 实例
func New(code int) *Result {
	return &Result{
		Code:    code,
		Message: Message(code),
		Detail:  make(map[string]string, 2),
	}
}

// NewWithDetail 声明一个带 Detail 内容的实例
func NewWithDetail(code int, detail map[string]string) *Result {
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
	return r.Code/Scale >= 400
}
