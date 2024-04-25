// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package html 提供输出 HTML 内容的解码函数
//
//	srv := server.New("", "", &server.Options{
//		Codec: web.New().AddMimetype("text/html", html.Marshal, html.Unmarshal, "")
//	})
//	tpl := template.ParseFiles(...)
//
//	func handle(ctx *web.Context) Responser {
//		obj := &struct{
//			HTMLName struct{} `html:"Object"`
//			Data string
//		}{}
//		return Object(200, obj, nil)
//	}
//
// 预定义的模板
//
// 框架本身提供了一些类型的定义，比如 [web.Problem]，当作为 html 输出时，其模板名称和数据为由
// [web.Problem.MarshalHTML] 方法返回，所以用户需要提供对应的模板定义。
// 如果用户还使用了 [server.RenderResponse]，那么也需要提供对应的模板定义。
package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"reflect"

	"github.com/issue9/mux/v8/header"
	"golang.org/x/text/message"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype"
)

const Mimetype = header.HTML

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
//   - 实现了 [Marshaler] 接口，则按 [Marshaler.MarshalHTML] 返回的查找模板名称；
//   - 其它普通对象，将获取对象的 HTMLName 的 struct tag，若不存在则直接采用类型名作为模板名；
//   - 其它情况下则是返回 [mimetype.ErrUnsupported]；
func Marshal(ctx *web.Context, v any) ([]byte, error) {
	if ctx == nil {
		panic("不支持该操作")
	}

	switch obj := v.(type) {
	case []byte:
		return obj, nil
	case string:
		return []byte(obj), nil
	default:
		return marshal(ctx, v)
	}
}

func marshal(ctx *web.Context, v any) ([]byte, error) {
	tt, found := ctx.Server().Vars().Load(viewContextKey)
	if !found {
		return nil, mimetype.ErrUnsupported()
	}
	vv := tt.(*view)

	tpl := vv.tpl
	if vv.dir {
		tag, _, _ := vv.b.Matcher().Match(ctx.LanguageTag())
		tagName := message.NewPrinter(tag, message.Catalog(vv.b)).Sprintf(tagKey)
		t, found := vv.dirTpls[tagName]
		if !found { // 理论上不可能出现此种情况，Match 必定会返回一个最相近的语种。
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

func Unmarshal(io.Reader, any) error { return mimetype.ErrUnsupported() }

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
