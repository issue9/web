// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package result 提供了一套用于描述向客户端反馈错误信息的机制。
package result

import (
	"net/http"
	"net/url"
	"strconv"

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
//    - field: username
//      message: 已经存在相同用户名
//    - field: username
//      message: 已经存在相同用户名
//
// FormData:
//  message=errormessage&code=4000001&detail.username=message&detail.username=message
type Result struct {
	XMLName struct{} `json:"-" xml:"result" yaml:"-"`

	// 当前的信息所对应的 HTTP 状态码
	Status int `json:"-" xml:"-" yaml:"-"`

	Message string    `json:"message" xml:"message,attr" yaml:"message"`
	Code    int       `json:"code" xml:"code,attr" yaml:"code"`
	Detail  []*detail `json:"detail,omitempty" xml:"field,omitempty" yaml:"detail,omitempty"`
}

type detail struct {
	Field   string `json:"field" xml:"name,attr" yaml:"field"`
	Message string `json:"message" xml:",chardata" yaml:"message"`
}

// SetDetail 设置详细的错误信息
//
// 会覆盖由 Add() 添加的内容
func (rslt *Result) SetDetail(fields map[string]string) *Result {
	rslt.Detail = make([]*detail, 0, len(fields))

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
	msg, found := messages[rslt.Code]
	if !found {
		logs.Error("不存在的错误码:", rslt.Code)
		ctx.Exit(http.StatusInternalServerError)
	}

	rslt.Status = msg.status
	rslt.Message = ctx.LocalePrinter.Sprintf(msg.message)
	ctx.Render(rslt.Status, rslt, nil)
}

// Exit 将当前的实例输出到客户端，并退出当前请求
func (rslt *Result) Exit(ctx *context.Context) {
	rslt.Render(ctx)
	ctx.Exit(0)
}

// MarshalForm 为 form.Marshaler 接口实现。用于将 result 对象转换成 form 数据格式
func (rslt *Result) MarshalForm() ([]byte, error) {
	vals := url.Values{}
	vals.Add("code", strconv.Itoa(rslt.Code))
	vals.Add("message", rslt.Message)

	for _, field := range rslt.Detail {
		vals.Add("detail."+field.Field, field.Message)
	}

	return []byte(vals.Encode()), nil
}
