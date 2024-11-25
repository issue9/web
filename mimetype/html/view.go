// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package html

import (
	"html/template"
	"io/fs"
	"maps"
	"slices"

	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

type contextType int

const viewContextKey contextType = 1

type view struct {
	tpls    map[language.Tag]*template.Template
	matcher language.Matcher
}

// Install 安装模板
//
// funcs 添加到当前模板系统的函数，除此之个，默认提供了以下两个函数：
//   - t 根据当前的语言（[web.Context.LanguageTag]）对参数进行翻译；
//   - tt 将内容翻译成指定语言，语言 ID 由第一个参数指定；
//
// localized 本地化 ID 与目录的映射关系，表示这些目录只解析至对应的 ID，
// 如果为空则相当于 {language.Und: "."}；
//
// fsys 表示模板目录。如果 localized 不为空，则只解析该参数中指定的目录，否则解析整个目录；
//
// NOTE: 可以多次调用，相同名称的模板会覆盖。
func Install(s web.Server, funcs template.FuncMap, localized map[language.Tag]string, glob string, fsys ...fs.FS) {
	var v *view
	if vv, ok := s.Vars().Load(viewContextKey); !ok {
		v = &view{}
		s.Vars().Store(viewContextKey, v)
	} else {
		v = vv.(*view)
	}

	if localized == nil {
		localized = map[language.Tag]string{language.Und: "."}
	}

	keys := slices.AppendSeq(slices.Collect(maps.Keys(v.tpls)), maps.Keys(localized))
	keys = sliceutil.Unique(keys, func(i, j language.Tag) bool { return i == j })
	v.matcher = language.NewMatcher(keys)

	for _, f := range fsys {
		install(s, v, funcs, localized, glob, f)
	}
}

func install(s web.Server, v *view, funcs template.FuncMap, localized map[language.Tag]string, glob string, fsys fs.FS) {
	tpls := make(map[language.Tag]*template.Template, len(localized))

	for t, name := range localized {
		sub, err := fs.Sub(fsys, name)
		if err != nil {
			panic(err)
		}

		tpl, found := v.tpls[t]
		if found {
			tpl = template.Must(tpl.Funcs(funcs).ParseFS(sub, glob))
		} else { // 首次添加
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

	tag, _, _ := v.matcher.Match(ctx.LanguageTag())
	tpl, found := v.tpls[tag]
	if !found && tag != language.Und {
		tpl = v.tpls[language.Und]
	}

	return tpl.Funcs(template.FuncMap{
		"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
	})
}
