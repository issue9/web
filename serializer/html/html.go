// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//	srv := NewServer()
//	tpl := template.ParseFiles(...)
//	srv.Mimetypes().Add("text/html", html.Marshal, html.Unmarshal)
//
//	func handle(ctx *web.Context) Responser {
//          obj := &struct{
//              HTMLName struct{} `html:"Object"`
//              Data string
//          }{}
//	    return Object(200, obj, nil)
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

// Marshaler 自定义 HTML 输出需要实现的接口
//
// 当前接口仅适用于由 [InstallView] 和 [InstallLocaleView] 管理的模板。
type Marshaler interface {
	// MarshalHTML 将对象转换成可用于模板的对象结构
	//
	// name 表示模板名称；
	// data 表示传递给该模板的数据；
	MarshalHTML() (name string, data any)
}

func newTpl(t *template.Template, name string, data any) *tpl {
	return &tpl{tpl: t, name: name, data: data}
}

// Marshal 针对 HTML 内容的解码实现
//
// 参数 v 可以是以下几种可能：
//   - string 或是 []byte 将内容作为 HTML 内容直接输出；
//   - 其它普通对象，将获取对象的 HTMLName 的 struct tag，若不存在则直接采用类型名作为模板名；
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
