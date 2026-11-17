// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"io"
	"io/fs"
	"iter"
	"slices"
	"strings"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/sliceutil"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/internal/qheader"
)

// Compressor 压缩算法的接口
type Compressor interface {
	// Name 算法的名称
	Name() string

	// NewDecoder 将 r 包装成为当前压缩算法的解码器
	NewDecoder(r io.Reader) (io.ReadCloser, error)

	// NewEncoder 将 w 包装成当前压缩算法的编码器
	NewEncoder(w io.Writer) (io.WriteCloser, error)
}

// MediaType 媒体类型需要提供的接口
type MediaType interface {
	// Name 获取媒体类型的名称
	Name() string

	// ProblemName 错误状态下的媒体类型
	ProblemName() string

	// Marshal 对内容进行编码
	Marshal(ctx *Context, v any) ([]byte, error)

	// Marshal 解码内容到 v
	Unmarshal(r io.Reader, v any) error
}

// Negotiation 提供与内容协商相关的功能
type Negotiation struct {
	disableCompress      bool
	compressions         []*compression
	acceptEncodingHeader string // 生成 AcceptEncoding 报头内容

	mediaTypes         []MediaType
	clientAcceptHeader string // 生成客户端的 Accept 报头内容
	serverAcceptHeader string // 生成服务端的 Accept 报头内容

	*catalog.Builder
	id       language.Tag
	printer  *message.Printer
	printers *ttlcache.Cache[language.Tag, *message.Printer]
	config   *config.Config

	acceptLanguage string // 根据内容生成 Accept-Language 报头
}

type compression struct {
	compressor     Compressor
	types          []string
	wildcard       bool     // 如果是通配符，则其它配置都将不启作用。
	wildcardSuffix []string // 是模糊类型的，比如 text/*，只有在 types 找不到时，才在此处查找。
}

const maxPrinters = 64 // 缓存的 [message.Printer] 数量上限

const maxPrinterIdle = 10 * time.Minute // 缓存项的空闲时长

// NewNegotiation 声明 [Negotiation] 对象
func NewNegotiation(id language.Tag, c *config.Config) *Negotiation {
	b := catalog.NewBuilder(catalog.Fallback(id))

	// 保证 b 中包含一条 id 语言的翻译项，
	// 这样可以始终让 Locale.Printer 的对象始终是有值的。
	if err := b.SetString(id, "_____", "_____"); err != nil {
		panic(err)
	}

	id, _, _ = b.Matcher().Match(id) // 从 b 中查找最合适的 tag

	return &Negotiation{
		compressions: make([]*compression, 0, 10),
		mediaTypes:   make([]MediaType, 0, 10),

		Builder: b,
		id:      id,
		config:  c,
		printer: message.NewPrinter(id, message.Catalog(b)),
		printers: ttlcache.New(
			ttlcache.WithCapacity[language.Tag, *message.Printer](maxPrinters),
			ttlcache.WithTTL[language.Tag, *message.Printer](maxPrinterIdle),
		),
		acceptLanguage: buildAcceptLanguage(b),
	}
}

func buildAcceptLanguage(ls catalog.Catalog) string {
	var s strings.Builder
	for index, v := range ls.Languages() {
		if index == 0 {
			s.WriteString(v.String())
		} else {
			s.WriteString(", ")
			s.WriteString(v.String())
		}
	}
	return s.String()
}

func (e *Negotiation) Config() *config.Config { return e.config }

// Language 返回默认的语言标签
func (e *Negotiation) Language() language.Tag { return e.id }

// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
//
// 根据 [Server.Config] 处理文件格式，如果文件格式不被 [Server.Config] 支持，将无法加载。
func (e *Negotiation) LoadMessages(glob string, fsys ...fs.FS) error {
	if err := locale.Load(e.config.Serializer(), e.Builder, glob, fsys...); err != nil {
		return err
	}

	e.acceptLanguage = buildAcceptLanguage(e.Builder)
	e.printers.DeleteAll()

	return nil
}

// LocalePrinter 最符合 [Locale.Language] 的 [message.LocalePrinter] 对象
func (e *Negotiation) LocalePrinter() *message.Printer { return e.printer }

// Sprintf 等同于 [Locale.Printer.Sprintf]
func (e *Negotiation) Sprintf(key string, v ...any) string {
	return e.LocalePrinter().Sprintf(key, v...)
}

// NewLocalePrinter 声明最符合 tag 的 [message.Printer] 对象
//
// NOTE: 每当给 [Locale.SetString]、[Locale.SetMacro] 和 [Locale.Set] 传递新的 [language.Tag]
// 值时，可能造成 NewLocalePrinter 相同的入参而返回不同的返回对象的情况。
func (e *Negotiation) NewLocalePrinter(id language.Tag) *message.Printer {
	if id == e.Language() {
		return e.LocalePrinter()
	}

	if item := e.printers.Get(id); item != nil {
		return item.Value()
	}

	// 从 l 中查找最合适的 tag
	id, _, _ = e.Matcher().Match(id)
	if item := e.printers.Get(id); item != nil {
		return item.Value()
	}

	p := message.NewPrinter(id, message.Catalog(e))
	e.printers.Set(id, p, maxPrinterIdle)
	return p
}

// SetString 添加新的翻译项
func (e *Negotiation) SetString(tag language.Tag, key, msg string) error {
	if err := e.Builder.SetString(tag, key, msg); err != nil {
		return err
	}
	e.printers.DeleteAll()
	return nil
}

// SetMacro 添加新的翻译项
func (e *Negotiation) SetMacro(tag language.Tag, name string, msg ...catalog.Message) error {
	if err := e.Builder.SetMacro(tag, name, msg...); err != nil {
		return err
	}
	e.printers.DeleteAll()
	return nil
}

// Set 添加新的翻译项
func (e *Negotiation) Set(tag language.Tag, key string, msg ...catalog.Message) error {
	if err := e.Builder.Set(tag, key, msg...); err != nil {
		return err
	}
	e.printers.DeleteAll()
	return nil
}

// AddCompressor 添加新的压缩算法
//
// t 表示适用的 content-type 类型，可以包含通配符，比如：
//
//	application/json
//	text/*
//	*
//
// 如果为空，则和 * 是相同的，表示匹配所有。
func (e *Negotiation) AddCompressor(c Compressor, t ...string) *Negotiation {
	e.compressions = append(e.compressions, buildCompression(c, t))

	names := make([]string, 0, len(e.compressions))
	for _, item := range e.compressions {
		names = append(names, item.compressor.Name())
	}
	names = sliceutil.Unique(names, func(i, j string) bool { return i == j })
	e.acceptEncodingHeader = strings.Join(names, ",")

	return e
}

func buildCompression(c Compressor, types []string) *compression {
	m := &compression{compressor: c}

	if len(types) == 0 {
		m.wildcard = true
		return m
	}

	m.types = make([]string, 0, len(types))
	m.wildcardSuffix = make([]string, 0, len(types))
	for _, c := range types {
		if c == "" {
			continue
		}

		if c == "*" {
			m.types = nil
			m.wildcardSuffix = nil
			m.wildcard = true
			return m
		}

		if c[len(c)-1] == '*' {
			m.wildcardSuffix = append(m.wildcardSuffix, c[:len(c)-1])
		} else {
			m.types = append(m.types, c)
		}
	}

	return m
}

// AddMediaType 添加媒体类型
//
// requestAccept 是否出现在客户端请求的 accept 报头中；
// responseAccept 是否出现在服务端返回的 accept 报头中；
func (e *Negotiation) AddMediaType(mt MediaType, requestAccept, responseAccept bool) *Negotiation {
	if mt == nil {
		panic("参数 mt 不能为空")
	}

	// 检测复复值
	if slices.ContainsFunc(e.mediaTypes, func(v MediaType) bool { return v.Name() == mt.Name() }) {
		panic(fmt.Sprintf("存在重复的项 %s", mt.Name()))
	}

	e.mediaTypes = append(e.mediaTypes, mt)

	if requestAccept {
		if e.clientAcceptHeader != "" {
			e.clientAcceptHeader += ", "
		}
		e.clientAcceptHeader += mt.Name()
	}

	if responseAccept {
		if e.serverAcceptHeader != "" {
			e.serverAcceptHeader += ", "
		}
		e.serverAcceptHeader += mt.Name()
	}

	return e
}

// 根据客户端的 Content-Encoding 报头对 r 进行包装
//
// name 编码名称，即 Content-Encoding 报头内容；
// r 为未解码的内容；
func (e *Negotiation) contentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if name == "" {
		return io.NopCloser(r), nil
	}

	if c, f := sliceutil.At(e.compressions, func(item *compression, _ int) bool { return item.compressor.Name() == name }); f {
		return c.compressor.NewDecoder(r)
	}
	return nil, NewError("not found compress for %s", name)
}

func unableParse(v string) error { return localeutil.Error("unable parse %s", v) }

// 根据客户端的 Accept-Encoding 报头选择是适合的压缩方法
//
// 如果返回的 c 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
// 如果 h 为空，那么将被设置为默认值 ＊；
func (e *Negotiation) acceptEncoding(contentType, h string) (c Compressor, err error) {
	// 规则如下：
	//  - 按 q 从大到小进行排序；
	//  - 忽略无法解析 q 的项；
	//  - 依次匹配，匹配到即返回；
	//  - 如果匹配到的项是 *，那么还要查看匹配到的项是否在 q=0 的项中，如果是则匹配下一个项；
	//
	// 如果传递的内容如下：
	//
	//	gzip;q=0.9,*;q=1
	//
	// 则因为 * 的 q 值比较高，而返回 * 匹配的内容
	//
	// */*;q=0 表示匹配所有未列出的编码

	if len(e.compressions) == 0 {
		return
	}

	accepts, deny, errs := qheader.ParseQHeader(h, "*")
	defer qheader.FreeItems(&accepts)

	if len(accepts) == 0 && len(deny) == 0 && len(errs) > 0 {
		return nil, unableParse(h)
	}

	indexes := e.getMatchCompresses(contentType)
	if len(indexes) == 0 { // 服务端未添加任何压缩算法
		if slices.Contains(deny, header.Identity) { // 且客户端要求必须要压缩
			return nil, localeutil.Error("client disabled identity")
		}
		return
	}

	hasAny := false
	for _, item := range accepts {
		if item.Value == "*" { // * 必然在切片的最后
			hasAny = true
			break
		}

		if item.Value == header.Identity { // 明确指定了不用压缩
			continue
		}

		for _, idx := range indexes {
			c := e.compressions[idx]
			if item.Value == c.compressor.Name() {
				return c.compressor, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 *，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, idx := range indexes {
			c := e.compressions[idx]
			if !denied(c.compressor.Name(), deny) {
				return c.compressor, nil
			}
		}
	}

	if slices.Contains(deny, header.Identity) { // 明确将 identity 设置为 0
		return nil, localeutil.Error("client disabled identity")
	}
	return nil, nil
}

func (e *Negotiation) getMatchCompresses(contentType string) []int {
	indexes := make([]int, 0, len(e.compressions))

LOOP:
	for index, c := range e.compressions {
		if c.wildcard {
			indexes = append(indexes, index)
			continue
		}

		for _, s := range c.types {
			if s == contentType {
				indexes = append(indexes, index)
				continue LOOP
			}
		}

		for _, p := range c.wildcardSuffix {
			if strings.HasPrefix(contentType, p) {
				indexes = append(indexes, index)
				continue LOOP
			}
		}
	}

	return indexes
}

// 从请求端提交的 Content-Type 报头中获取解码和字符集函数
//
// h 表示 Content-Type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (e *Negotiation) contentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mt, charset := qheader.ParseWithParam(h, "charset")

	item := e.searchFunc(func(s string) bool { return s == mt })
	if item == nil {
		return nil, nil, NewError("not found serialization function for %s", mt)
	}

	if charset == "" || charset == header.UTF8 {
		return item.Unmarshal, nil, nil
	}
	c, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return item.Unmarshal, c, nil
}

// 从请求端提交的 Accept 报头解析出所需要的解码函数
//
// 如果 h 为空，那么将被设置为默认值 ＊；
func (e *Negotiation) accept(h string) (MediaType, error) {
	// 规则如下：
	//  - 按 q 从大到小进行排序；
	//  - 忽略无法解析 q 的项；
	//  - 依次匹配，匹配到即返回；
	//  - 如果匹配到的项是 */*，那么还要查看匹配到的项是否在 q=0 的项中，如果是则匹配下一个项；
	//
	// 如果传递的内容如下：
	//
	//	application/json;q=0.9,*/*;q=1
	//
	// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
	//
	// */*;q=0 表示匹配所有未列出的编码

	accepts, deny, errs := qheader.ParseQHeader(h, "*/*")
	defer qheader.FreeItems(&accepts)

	if len(accepts) == 0 && len(deny) == 0 && len(errs) > 0 {
		return nil, unableParse(h)
	}

	hasAny := false
	for _, item := range accepts {
		if item.Value == "*/*" { // */* 必然在切片的最后
			hasAny = true
			break
		}

		for _, mt := range e.mediaTypes {
			if match(item.Value, mt) {
				return mt, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 */*，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, mt := range e.mediaTypes {
			if !denied(mt.Name(), deny) {
				return mt, nil
			}
		}
	}

	return nil, nil
}

func match(name string, mt MediaType) bool {
	return name == mt.Name() ||
		(strings.HasSuffix(name, "/*") && strings.HasPrefix(mt.Name(), name[:len(name)-1]))
}

// name 是否在 deny 中
func denied(name string, deny []string) bool {
	for _, d := range deny {
		if d == "*/*" || d == name {
			return true
		}
		if strings.HasSuffix(d, "/*") && strings.HasPrefix(name, d[:len(d)-1]) {
			return true
		}
	}
	return false
}

func (e *Negotiation) searchFunc(match func(string) bool) MediaType {
	item, _ := sliceutil.At(e.mediaTypes, func(i MediaType, _ int) bool {
		return match(i.Name()) || match(i.ProblemName())
	})
	return item
}

// CanCompress 当前是否拥有压缩功能
func (e *Negotiation) CanCompress() bool { return !e.disableCompress }

// SetCompress 设置压缩功能
//
// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
//
// NOTE: 仅对输出内容启作用，读取内容始终是按照提交的 Content-Encoding 指定算法进行解析。
func (e *Negotiation) SetCompress(enable bool) { e.disableCompress = !enable }

// MediaTypes 返回支持的所有媒体类型
//
// 键名和键值分别为正常和 problem 下的名称。
func (e *Negotiation) MediaTypes() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, v := range e.mediaTypes {
			if !yield(v.Name(), v.ProblemName()) {
				break
			}
		}
	}
}
