// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//	srv := NewServer()
//	tpl := template.ParseFiles(...)
//	srv.Mimetypes().Add("text/html", html.Marshal, html.Unmarshal)
//
//	func handle(ctx *web.Context) Responser {
//	    return Object(200, &struct{HTMLName struct{} `html:"Object"`}{}, nil)
//	}
package html

import (
	"bytes"
	"html/template"

	"github.com/issue9/web/serializer"
)

const Mimetype = "text/html"

type tpl struct {
	tpl  *template.Template
	name string // 模块名称
	data any    // 传递给模板的数据
}

func newTpl(t *template.Template, name string, data any) *tpl {
	return &tpl{tpl: t, name: name, data: data}
}

// Marshal 针对 HTML 内容的解码实现
//
// 参数 v 可以是以下几种可能：
//   - string 或是 []byte 将内容作为 HTML 内容直接输出；
//   - 其它情况下则是返回 [serializer.ErrUnsupported]；
func Marshal(v any) ([]byte, error) {
	switch obj := v.(type) {
	case *tpl:
		return obj.marshal()
	case []byte:
		return obj, nil
	case string:
		return []byte(obj), nil
	}
	return nil, serializer.ErrUnsupported
}

func Unmarshal([]byte, any) error { return serializer.ErrUnsupported }

func (t *tpl) marshal() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := t.tpl.ExecuteTemplate(w, t.name, t.data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
