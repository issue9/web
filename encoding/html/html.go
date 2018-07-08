// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package html 提供输出 HTML 内容的 encoding.MarshalFunc 函数。
//
//  tpl := template.ParseFiles(...)
//  mgr := html.New(tpl)
//  encoding.AddMarshal("text/html", mgr.Marshal)
//
//  func handle(w http.ResponseWriter, r *http.Request) {
//      ctx := web.New(w, r)
//      ctx.Render(html.Tpl("index", map[string]interface{}{...}))
//  }
package html

import (
	"bytes"
	"errors"
	"html/template"
)

// MimeType HTML 的 mimetype 值
const MimeType = "text/html"

var errUnsupported = errors.New("当前不支持该对象的解析")

// HTML 模板管理
type HTML struct {
	tpl *template.Template
}

// Template 传递给 encoding.MarshalFunc 的参数。
//
// 因为 encoding.MarshalFunc 限定了只能有一个参数，
// 而模板解析，除了传递的值，最起码还要一个模板名称，
// 所以采用 Template 作了一个简单的包装。
type Template struct {
	Name string // 模块名称
	Data interface{}
}

// Tpl 声明一个 *Template 变量。
//
// 其中 name 表示需要引用的模板名称，
// 而 data 则是传递给该模板的所有变量。
//
// 实际上就是
//  &Template {
//    Name: name,
//    Data: data,
//  }
// 的简单写法。
func Tpl(name string, data interface{}) *Template {
	return &Template{
		Name: name,
		Data: data,
	}
}

// New 声明 HTML 变量
func New(tpl *template.Template) *HTML {
	return &HTML{
		tpl: tpl,
	}
}

// Marshal 针对 HTML 内容的 MarshalFunc 实现
func (html *HTML) Marshal(v interface{}) ([]byte, error) {
	obj, ok := v.(*Template)
	if !ok {
		return nil, errUnsupported
	}

	w := new(bytes.Buffer)
	err := html.tpl.ExecuteTemplate(w, obj.Name, obj.Data)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}