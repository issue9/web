// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package html

import (
	"html/template"
	"io/fs"
	"maps"
	"slices"

	"golang.org/x/text/language"

	"github.com/issue9/web"
)

type contextType int

const viewContextKey contextType = 1

type view struct {
	localized map[language.Tag]string
	tpls      map[language.Tag]*template.Template
	matcher   language.Matcher
}

// Init 初始化 html 模板系统
//
// localized 模板的本地化目录列表。键值为目录名称，键名为该目录对应的本地化 ID。
// 将目录映射到 [language.Und]，表示该目录作为默认模板使用。
// 如果 localized 为空，则表示不按目录进行区分，将加载所有内容至一个模板实例中。
func Init(s web.Server, localized map[language.Tag]string) {
	if _, ok := s.Vars().Load(viewContextKey); ok {
		panic("已经初始化")
	}

	v := &view{localized: localized}

	if len(localized) > 0 {
		tags := slices.Collect(maps.Keys(localized))
		if slices.Index(tags, language.Und) < 0 {
			panic("必须指定 language.Und")
		}

		v.matcher = language.NewMatcher(tags)
	}

	s.Vars().Store(viewContextKey, v)
}

// Install 安装模板
//
// 提供了以下两个方法：
//   - t 根据当前的语言（[web.Context.LanguageTag]）对参数进行翻译；
//   - tt 将内容翻译成指定语言，语言 ID 由第一个参数指定；
//
// fsys 表示模板目录。该目录下应该包含 Init 中 localized 参数指定的所有目录，
// 否则会 panic。
//
// 通过此函数安装之后，可以正常输出以下内容：
//   - string 直接输出字符串；
//   - []byte 直接输出内容；
//   - Marshaler 将 [Marshaler.MarshalHTML] 返回内容作为输出内容；
//   - 其它结构体，尝试读取 XMLName 字段的 html struct tag 值作为模板名称进行查找；
//
// NOTE: 可以多次调用，相同名称的模板会覆盖。
func Install(s web.Server, funcs template.FuncMap, glob string, fsys ...fs.FS) {
	v, ok := s.Vars().Load(viewContextKey)
	if !ok {
		panic("未初始化")
	}

	vv := v.(*view)
	for _, f := range fsys {
		install(s, vv, funcs, glob, f)
	}
}

func install(s web.Server, v *view, funcs template.FuncMap, glob string, fsys fs.FS) {
	if len(v.localized) > 0 {
		instalLocalizedView(s, v, funcs, glob, fsys)
	} else {
		if len(v.tpls) == 0 { // 第一次调用
			f := getTranslateFuncMap(s)
			maps.Copy(f, funcs) // 用户定义的可覆盖系统自带的
			funcs = f
		}
		tpl := template.New(s.Name()).Funcs(funcs)
		template.Must(tpl.ParseFS(fsys, glob))
		s.Vars().Store(viewContextKey, &view{tpls: map[language.Tag]*template.Template{language.Und: tpl}})
	}
}

func instalLocalizedView(s web.Server, v *view, funcs template.FuncMap, glob string, fsys fs.FS) {
	tpls := make(map[language.Tag]*template.Template, len(v.localized))

	for t, name := range v.localized {
		sub, err := fs.Sub(fsys, name)
		if err != nil {
			panic(err)
		}

		tpl, found := v.tpls[t]
		if found { // 首次添加
			tpl = template.Must(tpl.Funcs(funcs).ParseFS(sub, glob))
		} else {
			f := getTranslateFuncMap(s)
			maps.Copy(f, funcs) // 用户定义的可覆盖系统自带的
			tpl = template.Must(template.New(name).Funcs(f).ParseFS(sub, glob))
		}

		tpls[t] = tpl
	}

	v.tpls = tpls
	s.Vars().Store(viewContextKey, v)
}

func getTranslateFuncMap(s web.Server) template.FuncMap {
	return template.FuncMap{
		"t": func(msg string, v ...any) string {
			return s.Locale().Printer().Sprintf(msg, v...)
		},

		"tt": func(tag, msg string, v ...any) string {
			return s.Locale().NewPrinter(language.MustParse(tag)).Sprintf(msg, v...)
		},
	}
}

// 生成适用当前会话的模板，这会添加一个当前语言的翻译方法 t
func (v *view) buildCurrentTpl(ctx *web.Context) *template.Template {
	if len(v.tpls) == 0 {
		return nil
	}

	var tpl *template.Template
	if len(v.localized) == 0 {
		tpl = v.tpls[language.Und]
	} else {
		tag, _, _ := v.matcher.Match(ctx.LanguageTag())
		tpl = v.tpls[tag]
	}

	return tpl.Funcs(template.FuncMap{
		"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
	})
}
