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
	// ctx 为上下文环境，本地化信息根据此参数中的 LanguageTag 确定；
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
}

// NewView 返回模板管理对象
//
// fsys 表示模板目录，如果为空则会采用 s 作为默认值；
// locale 是按目录对模板进行本地化分类。如果为 true
// 表示 fsys 下的一级目录为本地化的语言 ID；
func NewView(s *server.Server, fsys fs.FS, glob string, locale bool) View {
	if fsys == nil {
		fsys = s
	}

	funcs := template.FuncMap{
		"t": func(msg string, v ...any) string {
			return s.LocalePrinter().Sprintf(msg, v...)
		},
	}

	if locale {
		return newLocaleView(s, funcs, fsys, glob)
	}

	tpl := template.New(s.Name()).Funcs(funcs)
	template.Must(tpl.ParseFS(fsys, glob))
	return &tplView{template: tpl}
}

func newLocaleView(s *server.Server, funcs template.FuncMap, fsys fs.FS, glob string) View {
	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		panic(err)
	}

	b := catalog.NewBuilder()
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

		tpls[name] = template.Must(template.New(name).Funcs(funcs).ParseFS(sub, glob))
	}

	return &localeView{b: b, tpls: tpls}
}

func (v *localeView) View(ctx *server.Context, name string, data any) Marshaler {
	tag, _, _ := v.b.Matcher().Match(ctx.LanguageTag())
	tagName := message.NewPrinter(tag, message.Catalog(v.b)).Sprintf(tagKey)
	tpl := v.tpls[tagName]
	if tpl == nil {
		// TODO
		return nil
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
