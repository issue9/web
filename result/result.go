// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package result 提供了一套用于描述向客户端反馈错误信息的机制。
//
//
// 错误码
//
// 对于错误代码的定义是根据 HTTP 状态码进行分类的，
// 比如所有与 400 有关的错误信息，都是以 400 * 10N 为基数的；
// 而与验证有关的都是以 401 * 10N 为基数的。
package result

import (
	"net/http"

	"github.com/issue9/logs"
	"github.com/issue9/web/context"
)

// Result 定义了出错时，向客户端返回的结构体。支持以下格式：
//
// JSON:
//  {
//      'message': 'error message',
//      'code': 4000001,
//      'detail':[
//          {'field': 'username': 'message': '已经存在相同用户名'},
//          {'field': 'username': 'message': '已经存在相同用户名'},
//      ]
//  }
//
// XML:
//  <result code="400" message="error message">
//      <field name="username">已经存在相同用户名</field>
//      <field name="username">已经存在相同用户名</field>
//  </result>
//
// YAML:
//  message: 'error message'
//  code: 40000001
//  detail:
//    - field: usename
//      message: 已经存在相同用户名
//    - field: usename
//      message: 已经存在相同用户名
type Result struct {
	XMLName struct{} `json:"-" xml:"result" yaml:"-"`
	status  int      // 当前的信息所对应的 HTTP 状态码

	Message string    `json:"message" xml:"message,attr" yaml:"message"`
	Code    int       `json:"code" xml:"code,attr" yaml:"code"`
	Detail  []*detail `json:"detail,omitempty" xml:"field,omitempty" yaml:"detail,omitempty"`
}

type detail struct {
	Field   string `json:"field" xml:"name,attr" yaml:"field"`
	Message string `json:"message" xml:",chardata" yaml:"message"`
}

// New 声明一个新的 Result 实例
//
// code 表示错误代码；
// fields 为具体的错误信息，若没有，则为 nil；
func New(code int, fields map[string]string) *Result {
	msg, found := messages[code]
	if !found {
		logs.Error("不存在的错误码:", code)

		return &Result{
			Code:    -1,
			Message: "未知错误",
			status:  http.StatusInternalServerError,
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

// SetDetail 设置详细的错误信息
func (rslt *Result) SetDetail(fields map[string]string) *Result {
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

// HasDetail 是否包含详细的错误信息
func (rslt *Result) HasDetail() bool {
	return len(rslt.Detail) > 0
}

// Render 将当前的实例输出到客户端
func (rslt *Result) Render(ctx *context.Context) {
	ctx.Render(rslt.status, rslt, nil)
}
