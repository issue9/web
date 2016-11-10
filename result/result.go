// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// result 提供了一套用于描述向客户端反馈错误信息的机制。
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
	return r.Code/scale >= 400
}
