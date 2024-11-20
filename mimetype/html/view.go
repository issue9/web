// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package html

import (
	"fmt"
	"html/template"
	"io/fs"
	"maps"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
)

const tagKey = "view-locale-key"

type contextType int

const viewContextKey contextType = 1

type view struct {
	tpl *template.Template // 单目录模式下的模板

	dir     bool
	dirTpls map[string]*template.Template
	b       *catalog.Builder
}

// Init 初始化 html 模板系统
//
// dir 将 fsys 下的子目录作为本地化的 ID 对模板进行分类；
func Init(s web.Server, dir bool) {
	if _, ok := s.Vars().Load(viewContextKey); ok {
		panic("已经初始化")
	}

	v := &view{dir: dir}
	if dir {
		v.b = catalog.NewBuilder()
	}
	s.Vars().Store(viewContextKey, v)
}

// Install 安装模板
//
// 提供了以下两个方法：
//   - t 根据当前的语言（[web.Context.LanguageTag]）对参数进行翻译；
//   - tt 将内容翻译成指定语言，语言 ID 由第一个参数指定；
//
// fsys 表示模板目录；
//
// 通过此方法安装之后，可以正常处理用户提交的对象：
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
	if v.dir {
		instalDirView(s, v, funcs, glob, fsys)
	} else {
		if v.tpl == nil {
			f := getTranslateFuncMap(s)
			maps.Copy(f, funcs) // 用户定义的可覆盖系统自带的
			funcs = f
		}
		tpl := template.New(s.Name()).Funcs(funcs)
		template.Must(tpl.ParseFS(fsys, glob))
		s.Vars().Store(viewContextKey, &view{tpl: tpl})
	}
}

func instalDirView(s web.Server, v *view, funcs template.FuncMap, glob string, fsys fs.FS) {

	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		panic(err)
	}

	tpls := make(map[string]*template.Template, len(dirs))

	for _, dir := range dirs {
		name := dir.Name()

		tag, err := language.Parse(name)
		if err != nil {
			panic(fmt.Sprintf("无法将目录 %s 解析为本地化语言", name))
		}
		if err = v.b.SetString(tag, tagKey, name); err != nil {
			panic(err)
		}

		sub, err := fs.Sub(fsys, name)
		if err != nil {
			panic(err)
		}

		tpl, found := tpls[name]
		if found {
			tpl = template.Must(tpl.Funcs(funcs).ParseFS(sub, glob))
		} else {
			f := getTranslateFuncMap(s)
			maps.Copy(f, funcs) // 用户定义的可覆盖系统自带的
			tpl = template.Must(template.New(name).Funcs(f).ParseFS(sub, glob))
		}

		tpls[name] = tpl
	}

	s.Vars().Store(viewContextKey, &view{
		dir:     true,
		dirTpls: tpls,
		b:       v.b,
	})
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
	tpl := v.tpl
	if v.dir {
		tag, _, _ := v.b.Matcher().Match(ctx.LanguageTag())
		tagName := message.NewPrinter(tag, message.Catalog(v.b)).Sprintf(tagKey)
		t, found := v.dirTpls[tagName]
		if !found { // 理论上不可能出现此种情况。
			panic(fmt.Sprintf("未找到指定的模板 %s", tagName))
		}
		tpl = t
	}

	return tpl.Funcs(template.FuncMap{
		"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
	})
}
