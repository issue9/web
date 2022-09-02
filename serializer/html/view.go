// SPDX-License-Identifier: MIT

package html

import (
	"fmt"
	"html/template"
	"io/fs"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/server"
)

const tagKey = "view-locale-key"

// View 支持本地化的模板管理
type View interface {
	// View 返回输出 HTML 内容的对象
	//
	// 在模板中，用户可以使用 t 作为翻译函数，对内容进行翻译输出。
	// 其本地化的语言 ID 源自 ctx.LanguageTag。
	// t 的原型与模板内置的函数 printf 相同。
	// 同时还提供了 tt 输出指定语言的输出，相较于 t，tt 的第一个参数为语言 ID，
	// 比如 cmn-hans 等，要求必须能被 [language.Parse] 解析。
	//
	// name 为模板名称，data 为传递给模板的数据，
	// 这两个参数与 [template.Template.Execute] 中的相同。
	View(ctx *server.Context, name string, data any) Marshaler
}

type tplView struct {
	template *template.Template
}

type localeView struct {
	b    *catalog.Builder
	tpls map[string]*template.Template
	def  *template.Template
}

// NewView 返回本地化的模板
//
// 当前函数适合所有不同的本地化内容都在同一个模板中的，
// 通过翻译函数 t 输出各种语言的内容，模板中不能存在本地化相关的内容。
//
// fsys 表示模板目录，如果为空则会采用 s 作为默认值；
func NewView(s *server.Server, fsys fs.FS, glob string) View {
	fsys, funcs := initTpl(s, fsys)
	tpl := template.New(s.Name()).Funcs(funcs)
	template.Must(tpl.ParseFS(fsys, glob))
	return &tplView{template: tpl}
}

// NewLocaleView 声明目录形式的本地化模板
//
// 按目录名称加载各个本地化的模板，每个模板之间相互独立，模板内可以包含本地化相关的内容。
//
// fsys 表示模板目录，如果为空则会采用 s 作为默认值；
// def 表示默认的语言，必须是 fsys 下的目录名称；
func NewLocaleView(s *server.Server, fsys fs.FS, glob, def string) View {
	fsys, funcs := initTpl(s, fsys)

	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		panic(err)
	}

	b := catalog.NewBuilder()
	tpls := make(map[string]*template.Template, len(dirs))

	var defTpl *template.Template
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

		if name == def {
			defTpl = tpl
		}
	}

	if defTpl == nil {
		panic(fmt.Sprintf("指定的默认模板 %s 不存在", def))
	}

	return &localeView{b: b, tpls: tpls, def: defTpl}
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

func (v *localeView) View(ctx *server.Context, name string, data any) Marshaler {
	tag, _, _ := v.b.Matcher().Match(ctx.LanguageTag())
	tagName := message.NewPrinter(tag, message.Catalog(v.b)).Sprintf(tagKey)
	tpl := v.tpls[tagName]
	if tpl == nil {
		tpl = v.def
	}

	tpl = tpl.Funcs(template.FuncMap{
		"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
	})

	return Tpl(tpl, name, data)
}

func (v *tplView) View(ctx *server.Context, name string, data any) Marshaler {
	tpl := v.template.Funcs(template.FuncMap{
		"t": func(msg string, v ...any) string { return ctx.Sprintf(msg, v...) },
	})
	return Tpl(tpl, name, data)
}
