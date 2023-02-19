// SPDX-License-Identifier: MIT

package html

import (
	"fmt"
	"html/template"
	"io/fs"

	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/server"
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
//
// 通过此方法安装之后，可以正常处理用户提交的对象：
//   - string 直接输出字符串；
//   - []byte 直接输出内容；
//   - Marshaler 将 [Marshaler.MarshalHTML] 返回内容作为输出内容；
//   - 其它结构体，尝试读取 HTMLName 字段的 html struct tag 值作为模板名称进行查找；
//
// dir 表示是否以目录的形式组织本地化代码；
func InstallView(s *server.Server, dir bool, fsys fs.FS, glob string) {
	if dir {
		instalDirView(s, fsys, glob)
		return
	}

	fsys, funcs := initTpl(s, fsys)
	tpl := template.New(s.Name()).Funcs(funcs)
	template.Must(tpl.ParseFS(fsys, glob))

	s.Vars().Store(viewContextKey, &view{
		tpl: tpl,
	})
}

func instalDirView(s *server.Server, fsys fs.FS, glob string) {
	fsys, funcs := initTpl(s, fsys)

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

		tpl := template.Must(template.New(name).Funcs(funcs).ParseFS(sub, glob))
		tpls[name] = tpl
	}

	s.Vars().Store(viewContextKey, &view{
		dir:     true,
		dirTpls: tpls,
		b:       b,
	})
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
