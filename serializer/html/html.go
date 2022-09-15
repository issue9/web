// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//	srv := NewServer()
//	tpl := template.ParseFiles(...)
//	srv.Mimetypes().Add("text/html", html.Marshal, nil)
//
//	func handle(ctx *web.Context) Responser {
//	    return Object(200, html.Tpl(tpl, "index", map[string]any{...}), nil)
//	}
package html

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"

	"github.com/issue9/web/serializer"
)

const Mimetype = "text/html"

// Marshaler 实现输出 HTML 内容的接口
type Marshaler interface {
	MarshalHTML() ([]byte, error)
}

type tpl struct {
	// NOTE: 所有的字段值不能是可导出的。
	// 因为当用户的 accept 报头是 json 时，输出当前实例
	// 会使其所有的公开字段都被输出到客户端，存在一定的安全隐患。
	tpl  *template.Template
	name string // 模块名称
	data any    // 传递给模板的数据
}

// Tpl 将模板内容打包成 [Marshaler] 接口
//
// name 表示需要引用的模板名称；
// data 则是传递给该模板的所有变量；
func Tpl(t *template.Template, name string, data any) Marshaler {
	return &tpl{
		tpl:  t,
		name: name,
		data: data,
	}
}

// Marshal 针对 HTML 内容的解码实现
//
// 参数 v 可以是以下几种可能：
//   - Marshaler 接口；
//   - string 或是 []byte 将内容作为 HTML 内容直接输出；
//   - 其它情况下则是返回 [serializer.ErrUnsupported]；
func Marshal(v any) ([]byte, error) {
	switch obj := v.(type) {
	case Marshaler:
		return obj.MarshalHTML()
	case []byte:
		return obj, nil
	case string:
		return []byte(obj), nil
	}
	return nil, serializer.ErrUnsupported
}

func Unmarshal([]byte, any) error { return serializer.ErrUnsupported }

func (t *tpl) MarshalHTML() ([]byte, error) {
	w := new(bytes.Buffer)
	if err := t.tpl.ExecuteTemplate(w, t.name, t.data); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (t *tpl) MarshalJSON() ([]byte, error) { return json.Marshal(t.data) }

func (t *tpl) MarshalXML(e *xml.Encoder, s xml.StartElement) error { return e.EncodeElement(t.data, s) }
