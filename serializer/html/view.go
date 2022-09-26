// SPDX-License-Identifier: MIT

package html

import (
	"fmt"
	"html/template"
	"io/fs"
	"reflect"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/server"
)

const tagKey = "view-locale-key"

// InstallView 返回本地化的模板
//
// 适合所有不同的本地化内容都在同一个模板中的，
// 通过翻译函数 t 输出各种语言的内容，模板中不能存在本地化相关的内容。
//
// 提供了以下两个方法：
//   - t 根据当前的语言（[server.Context.LanguageTag]）对参数进行翻译；
//   - tt 将内容翻译成指定语言，语言 ID 由第一个参数指定；
//
// fsys 表示模板目录，如果为空则会采用 s 作为默认值；
func InstallView(s *server.Server, fsys fs.FS, glob string) {
	fsys, funcs := initTpl(s, fsys)
	tpl := template.New(s.Name()).Funcs(funcs)
	template.Must(tpl.ParseFS(fsys, glob))

	s.OnMarshal(Mimetype, func(ctx *server.Context, a any) any {
		tpl := tpl.Funcs(template.FuncMap{
			"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
		})

		name, v := getName(a)
		return newTpl(tpl, name, v)
	}, nil)
}

// InstallLocaleView 声明目录形式的本地化模板
//
// 按目录名称加载各个本地化的模板，每个模板之间相互独立，模板内可以包含本地化相关的内容。
//
// 提供了以下两个方法：
//   - t 根据当前的语言（[server.Context.LaguageTag]）对参数进行翻译；
//   - tt 将内容翻译成指定语言，语言 ID 由第一个参数指定；
//
// fsys 表示模板目录，如果为空则会采用 s 作为默认值；
func InstallLocaleView(s *server.Server, fsys fs.FS, glob string) {
	fsys, funcs := initTpl(s, fsys)

	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		panic(err)
	}

	b := catalog.NewBuilder(catalog.Fallback(s.LanguageTag()))
	tpls := make(map[string]*template.Template, len(dirs))

	for _, dir := range dirs {
		name := dir.Name()

		tag, err := language.Parse(name)
		if err != nil {
			panic(fmt.Sprintf("无法将目录 %s 解析为本地化语言", name))
		}
		if err = b.SetString(tag, tagKey, name); err != nil {
			panic(err)
		}

		sub, err := fs.Sub(fsys, name)
		if err != nil {
			panic(err)
		}

		tpl := template.Must(template.New(name).Funcs(funcs).ParseFS(sub, glob))
		tpls[name] = tpl
	}

	installLocaleView(s, b, tpls)
}

func initTpl(s *server.Server, fsys fs.FS) (fs.FS, template.FuncMap) {
	if fsys == nil {
		fsys = s
	}

	funcs := template.FuncMap{
		"t": func(msg string, v ...any) string {
			return s.LocalePrinter().Sprintf(msg, v...)
		},

		"tt": func(tag, msg string, v ...any) string {
			return s.NewPrinter(language.MustParse(tag)).Sprintf(msg, v...)
		},
	}

	return fsys, funcs
}

func installLocaleView(s *server.Server, b *catalog.Builder, tpls map[string]*template.Template) {
	s.OnMarshal(Mimetype, func(ctx *server.Context, a any) any {
		tag, _, _ := b.Matcher().Match(ctx.LanguageTag())
		tagName := message.NewPrinter(tag, message.Catalog(b)).Sprintf(tagKey)
		tpl, found := tpls[tagName]
		if !found { // 理论上不可能出现此种情况，Match 必定返回一个最相近的语种。
			panic(fmt.Sprintf("未找到指定的模板 %s", tagName))
		}

		tpl = tpl.Funcs(template.FuncMap{
			"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
		})

		name, v := getName(a)
		return newTpl(tpl, name, v)
	}, nil)
}

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
