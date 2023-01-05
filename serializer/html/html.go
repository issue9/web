// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//		srv := NewServer()
//		tpl := template.ParseFiles(...)
//		srv.Mimetypes().Add("text/html", html.Marshal, html.Unmarshal)
//
//		func handle(ctx *web.Context) Responser {
//	         obj := &struct{
//	             HTMLName struct{} `html:"Object"`
//	             Data string
//	         }{}
//		    return Object(200, obj, nil)
//		}
package html

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"

	"golang.org/x/text/message"

	"github.com/issue9/web/serializer"
	"github.com/issue9/web/server"
)

const Mimetype = "text/html"

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

// Marshal 针对 HTML 内容的解码实现
//
// 参数 v 可以是以下几种可能：
//   - string 或是 []byte 将内容作为 HTML 内容直接输出；
//   - 其它普通对象，将获取对象的 HTMLName 的 struct tag，若不存在则直接采用类型名作为模板名；
//   - 其它情况下则是返回 [server.ErrUnsupported]；
func Marshal(ctx *server.Context, v any) ([]byte, error) {
	switch obj := v.(type) {
	case []byte:
		return obj, nil
	case string:
		return []byte(obj), nil
	default:
		return marshal(ctx, v)
	}
}

func marshal(ctx *server.Context, v any) ([]byte, error) {
	tt, found := ctx.Server().Vars().Load(viewContextKey)
	if !found {
		return nil, serializer.ErrUnsupported()
	}
	vv := tt.(*view)

	tpl := vv.tpl
	if vv.dir {
		tag, _, _ := vv.b.Matcher().Match(ctx.LanguageTag())
		tagName := message.NewPrinter(tag, message.Catalog(vv.b)).Sprintf(tagKey)
		t, found := vv.dirTpls[tagName]
		if !found { // 理论上不可能出现此种情况，Match 必定返回一个最相近的语种。
			panic(fmt.Sprintf("未找到指定的模板 %s", tagName))
		}
		tpl = t
	}

	tpl = tpl.Funcs(template.FuncMap{
		"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
	})

	name, v := getName(v)

	w := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(w, name, v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func Unmarshal([]byte, any) error { return serializer.ErrUnsupported() }

func getName(v any) (string, any) {
	if m, ok := v.(Marshaler); ok {
		return m.MarshalHTML()
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		if name := rt.Name(); name != "" {
			return name, v
		}
		panic(fmt.Sprintf("text/html 不支持输出当前类型 %s", rt.Kind()))
	}

	field, found := rt.FieldByName("HTMLName")
	if !found {
		return rt.Name(), v
	}

	tag := field.Tag.Get("html")
	if tag == "" {
		tag = rt.Name()
	}
	return tag, v
}
