// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"io"
	"io/fs"
	"iter"
	"net/http"
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

	builder              *catalog.Builder
	lang                 language.Tag
	printer              *message.Printer
	printers             *ttlcache.Cache[language.Tag, *message.Printer]
	config               *config.Config
	acceptLanguageHeader string // 根据 builder 加载的语言文件，生成服务端的 Accept-Language 报头
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
//
// conf 为加载配置文件的对象，[Negotiation.LoadMessags] 使用该值加载本地化的语言文件。
func NewNegotiation(lang language.Tag, conf *config.Config) *Negotiation {
	b := catalog.NewBuilder(catalog.Fallback(lang))

	// 保证 b 中包含一条 id 语言的翻译项，这样可以始终让 Locale.Printer 的对象始终是有值的。
	if err := b.SetString(lang, "_____", "_____"); err != nil {
		panic(err)
	}

	lang, _, _ = b.Matcher().Match(lang) // 从 b 中查找最合适的 tag

	return &Negotiation{
		compressions: make([]*compression, 0, 10),
		mediaTypes:   make([]MediaType, 0, 10),

		builder: b,
		lang:    lang,
		config:  conf,
		printer: message.NewPrinter(lang, message.Catalog(b)),
		printers: ttlcache.New(
			ttlcache.WithCapacity[language.Tag, *message.Printer](maxPrinters),
			ttlcache.WithTTL[language.Tag, *message.Printer](maxPrinterIdle),
		),
		acceptLanguageHeader: buildAcceptLanguage(b),
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

// Language 返回默认的语言标签
func (nego *Negotiation) Language() language.Tag { return nego.lang }

func (nego *Negotiation) Catalog() *catalog.Builder { return nego.builder }

// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
//
// 根据 [Server.Config] 处理文件格式，如果文件格式不被 [Server.Config] 支持，将无法加载。
func (nego *Negotiation) LoadMessages(glob string, fsys ...fs.FS) error {
	if err := locale.Load(nego.config.Serializer(), nego.builder, glob, fsys...); err != nil {
		return err
	}

	nego.acceptLanguageHeader = buildAcceptLanguage(nego.Catalog())
	nego.printers.DeleteAll()

	return nil
}

// LocalePrinter 最符合 [Locale.Language] 的 [message.LocalePrinter] 对象
func (nego *Negotiation) LocalePrinter() *message.Printer { return nego.printer }

// Sprintf 等同于 [Locale.Printer.Sprintf]
func (nego *Negotiation) Sprintf(key string, v ...any) string {
	return nego.LocalePrinter().Sprintf(key, v...)
}

// NewLocalePrinter 声明最符合 tag 的 [message.Printer] 对象
//
// NOTE: 每当给 [Locale.SetString]、[Locale.SetMacro] 和 [Locale.Set] 传递新的 [language.Tag]
// 值时，可能造成 NewLocalePrinter 相同的入参而返回不同的返回对象的情况。
func (nego *Negotiation) NewLocalePrinter(id language.Tag) *message.Printer {
	if id == nego.Language() {
		return nego.LocalePrinter()
	}

	if item := nego.printers.Get(id); item != nil {
		return item.Value()
	}

	// 从 l 中查找最合适的 tag
	id, _, _ = nego.Catalog().Matcher().Match(id)
	if item := nego.printers.Get(id); item != nil {
		return item.Value()
	}

	p := message.NewPrinter(id, message.Catalog(nego.Catalog()))
	nego.printers.Set(id, p, maxPrinterIdle)
	return p
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
func (nego *Negotiation) AddCompressor(c Compressor, t ...string) *Negotiation {
	nego.compressions = append(nego.compressions, buildCompression(c, t))

	names := make([]string, 0, len(nego.compressions))
	for _, item := range nego.compressions {
		names = append(names, item.compressor.Name())
	}
	names = sliceutil.Unique(names, func(i, j string) bool { return i == j })
	nego.acceptEncodingHeader = strings.Join(names, ",")

	return nego
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
func (nego *Negotiation) AddMediaType(mt MediaType, requestAccept, responseAccept bool) *Negotiation {
	if mt == nil {
		panic("参数 mt 不能为空")
	}

	// 检测复复值
	if slices.ContainsFunc(nego.mediaTypes, func(v MediaType) bool { return v.Name() == mt.Name() }) {
		panic(fmt.Sprintf("存在重复的项 %s", mt.Name()))
	}

	nego.mediaTypes = append(nego.mediaTypes, mt)

	if requestAccept {
		if nego.clientAcceptHeader != "" {
			nego.clientAcceptHeader += ", "
		}
		nego.clientAcceptHeader += mt.Name()
	}

	if responseAccept {
		if nego.serverAcceptHeader != "" {
			nego.serverAcceptHeader += ", "
		}
		nego.serverAcceptHeader += mt.Name()
	}

	return nego
}

// 根据客户端的 Content-Encoding 报头对 r 进行包装
//
// name 编码名称，即 Content-Encoding 报头内容；
// r 为未解码的内容；
func (nego *Negotiation) contentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if name == "" {
		return io.NopCloser(r), nil
	}

	if c, f := sliceutil.At(nego.compressions, func(item *compression, _ int) bool { return item.compressor.Name() == name }); f {
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
func (nego *Negotiation) acceptEncoding(contentType, h string) (c Compressor, err error) {
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

	if len(nego.compressions) == 0 {
		return
	}

	accepts, deny, errs := qheader.ParseQHeader(h, "*")
	defer qheader.FreeItems(&accepts)

	if len(accepts) == 0 && len(deny) == 0 && len(errs) > 0 {
		return nil, unableParse(h)
	}

	indexes := nego.getMatchCompresses(contentType)
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
			c := nego.compressions[idx]
			if item.Value == c.compressor.Name() {
				return c.compressor, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 *，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, idx := range indexes {
			c := nego.compressions[idx]
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

func (nego *Negotiation) getMatchCompresses(contentType string) []int {
	indexes := make([]int, 0, len(nego.compressions))

LOOP:
	for index, c := range nego.compressions {
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
func (nego *Negotiation) contentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mt, charset := qheader.ParseWithParam(h, "charset")

	item := nego.searchFunc(func(s string) bool { return s == mt })
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
func (nego *Negotiation) accept(h string) (MediaType, error) {
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

		for _, mt := range nego.mediaTypes {
			if match(item.Value, mt) {
				return mt, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 */*，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, mt := range nego.mediaTypes {
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

func (nego *Negotiation) searchFunc(match func(string) bool) MediaType {
	item, _ := sliceutil.At(nego.mediaTypes, func(i MediaType, _ int) bool {
		return match(i.Name()) || match(i.ProblemName())
	})
	return item
}

// CanCompress 当前是否拥有压缩功能
func (nego *Negotiation) CanCompress() bool { return !nego.disableCompress }

// SetCompress 设置压缩功能
//
// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
//
// NOTE: 仅对输出内容启作用，读取内容始终是按照提交的 Content-Encoding 指定算法进行解析。
func (nego *Negotiation) SetCompress(enable bool) { nego.disableCompress = !enable }

// MediaTypes 返回支持的所有媒体类型
//
// 键名和键值分别为正常和 problem 下的名称。
func (nego *Negotiation) MediaTypes() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, v := range nego.mediaTypes {
			if !yield(v.Name(), v.ProblemName()) {
				break
			}
		}
	}
}

func (nego *Negotiation) setVary(h http.Header) {
	h.Add(header.Vary, header.Accept)
	h.Add(header.Vary, header.AcceptLanguage)
	h.Add(header.Vary, header.AcceptEncoding)
}
