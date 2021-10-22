// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//  mt := content.NewContent()
//  tpl := template.ParseFiles(...)
//  mt.Add("text/html", html.Marshal, nil)
//
//  func handle(ctx *web.Context) Responser {
//      return Object(200, html.Tpl(tpl, "index", map[string]interface{}{...}), nil)
//  }
package html

import (
	"bytes"
	"html/template"
	"io"

	"github.com/issue9/web/serialization"
)

// Mimetype HTML 的 mimetype 值
const Mimetype = "text/html"

// Template 传递给 Marshal 的参数
type Template struct {
	Template *template.Template
	Name     string      // 模块名称
	Data     interface{} // 传递给模板的数据
}

// Tpl 声明一个 *Template 变量
//
// 其中 name 表示需要引用的模板名称，
// 而 data 则是传递给该模板的所有变量。
func Tpl(tpl *template.Template, name string, data interface{}) *Template {
	return &Template{
		Template: tpl,
		Name:     name,
		Data:     data,
	}
}

// Marshal 针对 HTML 内容的解码实现
//
// 参数 v 限定为 *Template 类型，否则将返回错误。
func Marshal(v interface{}) ([]byte, error) {
	obj, ok := v.(*Template)
	if !ok {
		return nil, serialization.ErrUnsupported
	}

	w := new(bytes.Buffer)
	if err := obj.executeTemplate(w); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (t *Template) executeTemplate(w io.Writer) error {
	return t.Template.ExecuteTemplate(w, t.Name, t.Data)
}
