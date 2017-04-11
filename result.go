// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/logs"
)

// Result 提供了一套用于描述向客户端反馈错误信息的机制。
//
// 对于错误代码的定义是根据 HTTP 状态码进行分类的，
// 比如所有与 400 有关的错误信息，都是以 400 * 10N 为基数的；
// 而与验证有关的都是以 401 * 10N 为基数的。
//
// 在 Result.IsError() 为 true 的情况下，也可以将其当作 error 使用。
//
// Result 可以向用户输出以下格式的数据类型：
// JSON
//  {
//      'message': 'error message',
//      'code': 4000001,
//      'detail':[
//          {'field': 'username': 'message': '已经存在相同用户名'},
//          {'field': 'username': 'message': '已经存在相同用户名'},
//      ]
//  }
// 或是 XML
//  <result code="400" message="error message">
//      <field name="username">已经存在相同用户名</field>
//      <field name="username">已经存在相同用户名</field>
//  </result>
type Result struct {
	XMLName struct{} `json:"-" xml:"result"`
	status  int      // 当前的信息所对应的 HTTP 状态码

	Message string    `json:"message" xml:"message,attr"`
	Code    int       `json:"code" xml:"code,attr"`
	Detail  []*detail `json:"detail,omitempty" xml:"field,omitempty"`
}

type detail struct {
	Field   string `json:"field" xml:"name,attr"`
	Message string `json:"message" xml:",chardata"`
}

// NewResult 声明一个新的 Result 实例
//
// code 表示错误代码；
// fields 为具体的错误信息；
func NewResult(code int, fields map[string]string) *Result {
	msg, err := getMessage(code)
	if err != nil {
		logs.Error(err)

		return &Result{
			Code:    -1,
			Message: msg.message,
			status:  msg.status,
		}
	}

	rslt := &Result{
		Code:    code,
		Message: msg.message,
		status:  msg.status,
		Detail:  make([]*detail, 0, len(fields)),
	}

	for k, v := range fields {
		rslt.Add(k, v)
	}

	return rslt
}

// Add 添加一条详细的错误信息。
//
// 若 field 与已有的同名，会出现多条同名记录。
func (rslt *Result) Add(field, message string) *Result {
	rslt.Detail = append(rslt.Detail, &detail{Field: field, Message: message})
	return rslt
}

// Error error 接口
//
// 具体是否为一个 error 接口，还需要查看 IsError 是否为 true
func (rslt *Result) Error() string {
	return rslt.Message
}

// HasDetail 是否包含详细的错误信息
func (rslt *Result) HasDetail() bool {
	return len(rslt.Detail) > 0
}

// IsError 当将 Result 当作 error 实例来用时，需要判断此值是否为 true。
func (rslt *Result) IsError() bool {
	return rslt.status >= http.StatusBadRequest
}

// Status 获取与其相对的 HTTP 状态码
func (rslt *Result) Status() int {
	return rslt.status
}

// Render 将当前的实例输出到客户端
func (rslt *Result) Render(ctx *Context) {
	ctx.Render(rslt.status, rslt, nil)
}
