// SPDX-License-Identifier: MIT

package serialization

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type (
	// Locale 提供了从文件中加载本地化信息的方法
	Locale struct {
		b *catalog.Builder
		f *Files
	}
)

// NewLocale 返回 Locale 实例
//
// f 表示用于加载本地化文件的序列化方法，根据文件扩展名在 f 中查找相应的序列化方法；
// 加载后的内容被应用在 b 之上。
func NewLocale(b *catalog.Builder, f *Files) *Locale { return &Locale{b: b, f: f} }

// Files 返回用于序列化文件的实例
func (l *Locale) Files() *Files { return l.f }

// Builder 返回本地化操作的相关接口
func (l *Locale) Builder() *catalog.Builder { return l.b }

func (l *Locale) Printer(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(l.Builder()))
}

// LoadFile 从文件中加载本地化内容
func (l *Locale) LoadFile(glob string) error {
	matchs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, f := range matchs {
		_, u := l.Files().searchByExt(f)
		if u == nil {
			return fmt.Errorf("未找到适合 %s 的函数", f)
		}

		if err := localeutil.LoadMessageFromFile(l.b, f, u); err != nil {
			return err
		}

	}

	return nil
}

// LoadFileFS 从文件中加载本地化内容
func (l *Locale) LoadFileFS(fsys fs.FS, glob string) error {
	matchs, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	for _, f := range matchs {
		_, u := l.Files().searchByExt(f)
		if u == nil {
			return fmt.Errorf("未找到适合 %s 的函数", f)
		}

		if err := localeutil.LoadMessageFromFS(l.b, fsys, f, u); err != nil {
			return err
		}
	}

	return nil
}