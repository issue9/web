// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package negotiation 提供与内容协商相关的功能
//
// 当前包为了与主包解耦，将 [Context] 抽象为为泛型 T。
package negotiation

import (
	"io"
	"io/fs"
	"net/http"

	"github.com/issue9/config"
	"github.com/issue9/mux/v10/header"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/locale"
)

// MarshalFunc 序列化函数原型
//
// NOTE: MarshalFunc 的作用是输出内容，所以在实现中不能调用 [Context.Render] 等输出方法。
//
// NOTE: 不采用流的方式处理数据的原因是因为：编码过程中可能会出错，
// 此时需要修改状态码，流式的因为有内容输出，状态码也已经固定，无法修改。
type MarshalFunc[T any] func(T, any) ([]byte, error)

// UnmarshalFunc 反序列化函数原型
//
// NOTE: 参数 [io.Reader] 必定不会为空。
type UnmarshalFunc func(io.Reader, any) error

// Negotiation 提供与内容协商相关的功能
type Negotiation[T any] struct {
	l *locale.Locale // 本地化

	disableCompress      bool
	compressions         []*compression
	acceptEncodingHeader string // 生成 AcceptEncoding 报头内容

	mediaTypes        []*MediaType[T]
	acceptHeader      string // 生成的 Accept 报头内容
	acceptPatchHeader string // 生成的 Accept-Patch 报头内容
	acceptQueryHeader string // 生成的 Accept-Query 报头内容
}

// New 声明 [Negotiation] 对象
//
// conf 为加载配置文件的对象，[Negotiation.LoadMessages] 使用该值加载本地化的语言文件。
func New[T any](lang language.Tag, conf *config.Config) *Negotiation[T] {
	return &Negotiation[T]{
		compressions: make([]*compression, 0, 10),
		mediaTypes:   make([]*MediaType[T], 0, 10),
		l:            locale.New(lang, conf),
	}
}

// Language 返回默认的语言标签
func (nego *Negotiation[T]) Language() language.Tag { return nego.l.ID() }

func (nego *Negotiation[T]) Catalog() *catalog.Builder { return nego.l.Catalog() }

func (nego *Negotiation[T]) AcceptLanguageHeader() string { return nego.l.AcceptLanguageHeader() }

func (nego *Negotiation[T]) Config() *config.Config { return nego.l.Config() }

// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
//
// 根据 [Server.Config] 处理文件格式，如果文件格式不被支持，将无法加载。
func (nego *Negotiation[T]) LoadMessages(glob string, fsys ...fs.FS) error {
	return nego.l.LoadMessages(glob, fsys...)
}

// LocalePrinter 最符合 [Locale.Language] 的 [message.LocalePrinter] 对象
func (nego *Negotiation[T]) LocalePrinter() *message.Printer { return nego.l.Printer() }

// Sprintf 等同于 [Locale.Printer.Sprintf]
func (nego *Negotiation[T]) Sprintf(key string, v ...any) string {
	return nego.LocalePrinter().Sprintf(key, v...)
}

// NewLocalePrinter 声明最符合 tag 的 [message.Printer] 对象
//
// NOTE: 每当更新 [Negotiation.Catalog] 时，可能造成 NewLocalePrinter 相同的入参而返回不同的返回对象的情况。
func (nego *Negotiation[T]) NewLocalePrinter(id language.Tag) *message.Printer {
	return nego.l.NewPrinter(id)
}

// 根据 Accept-Language 报头获取最适合的语言
func (nego *Negotiation[T]) AcceptLanguage(h string) language.Tag {
	if h == "" {
		return nego.Language()
	}
	tag, _ := language.MatchStrings(nego.Catalog().Matcher(), h)
	return tag
}

func (nego *Negotiation[T]) SetVary(h http.Header) {
	h.Add(header.Vary, header.Accept)
	h.Add(header.Vary, header.AcceptLanguage)
	h.Add(header.Vary, header.AcceptEncoding)
}
