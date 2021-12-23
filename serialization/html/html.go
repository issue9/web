// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//  srv := NewServer()
//  tpl := template.ParseFiles(...)
//  srv.Mimetypes().Add("text/html", html.Marshal, nil)
//
//  func handle(ctx *web.Context) Responser {
//      return Object(200, html.Tpl(tpl, "index", map[string]interface{}{...}), nil)
//  }
package html

import (
	"bytes"
	"html/template"

	"github.com/issue9/web/serialization"
)

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
// 参数 v 可以是以下几种类型：
//  - string 或是 []byte 将内容作为 HTML 内容直接输出
//  - *Template 编译模板内容并输出
//  - 其它情况下则是返回错误。
func Marshal(v interface{}) ([]byte, error) {
	switch obj := v.(type) {
	case *Template:
		return obj.executeTemplate()
	case []byte:
		return obj, nil
	case string:
		return []byte(obj), nil
	}
	return nil, serialization.ErrUnsupported
}

func (t *Template) executeTemplate() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := t.Template.ExecuteTemplate(w, t.Name, t.Data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
