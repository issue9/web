// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的 content.MarshalFunc 函数
//
//  mt := content.NewMimetypes()
//  tpl := template.ParseFiles(...)
//  mgr := html.New(tpl)
//  mt.Add("text/html", mgr.Marshal, nil)
//
//  func handle(ctx *web.Context) {
//      ctx.Render(200, html.Tpl("index", map[string]interface{}{...}), nil)
//  }
package html

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/issue9/web/content"
)

// Mimetype HTML 的 mimetype 值
const Mimetype = "text/html"

var errUnsupported = errors.New("当前不支持该对象的解析")

// HTML 模板管理
type HTML struct {
	tpl *template.Template
}

// Template 传递给 content.MarshalFunc 的参数
//
// 因为 content.MarshalFunc 限定了只能有一个参数，
// 而模板解析，除了传递的值，最起码还要一个模板名称，
// 所以采用 Template 作了一个简单的包装。
type Template struct {
	Name string // 模块名称
	Data interface{}
}

// Tpl 声明一个 *Template 变量
//
// 其中 name 表示需要引用的模板名称，
// 而 data 则是传递给该模板的所有变量。
func Tpl(name string, data interface{}) *Template {
	return &Template{
		Name: name,
		Data: data,
	}
}

// New 声明 HTML 变量
//
// tpl 可以为空，通过之后的 SetTemplate 再次指定
func New(tpl *template.Template) *HTML {
	return &HTML{
		tpl: tpl,
	}
}

// SetTemplate 修改模板内容
func (html *HTML) SetTemplate(tpl *template.Template) {
	html.tpl = tpl
}

// Marshal 针对 HTML 内容的 content.MarshalFunc 实现
//
// 参数 v 限定为 *Template 类型，否则将返回错误。
func (html *HTML) Marshal(v interface{}) ([]byte, error) {
	if v == content.Nil {
		return nil, nil
	}

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
