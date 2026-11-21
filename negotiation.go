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

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/internal/qheader"
)

// Compressor 压缩算法的接口
//
// 实现者应该是一个无状态的对象。
type Compressor interface {
	// Name 算法的名称
	Name() string

	// NewDecoder 将 r 包装成为当前压缩算法的解码器
	NewDecoder(r io.Reader) (io.ReadCloser, error)

	// NewEncoder 将 w 包装成当前压缩算法的编码器
	NewEncoder(w io.Writer) (io.WriteCloser, error)
}

// Codec 媒体类型需要提供的编码接口
//
// 实现者应该是一个无状态的对象。
type Codec interface {
	// Marshal 对内容进行编码
	Marshal(ctx *Context, v any) ([]byte, error)

	// Marshal 解码内容到 v
	Unmarshal(r io.Reader, v any) error
}

// MediaType 媒体类型
type MediaType struct {
	Codec
	name, problem, method string
	kind                  PatchKind
}

type PatchKind = int8

const (
	PatchKindNone PatchKind = iota
	PatchKindPatcher
	PatchKindMerger
)

// NewMediaType 创建新的媒体类型
//
// name 媒体类型的名称；
// problem 错误状态下的媒体类型；
// method 绑定的方法；
// pathKind 在 method 的值为 [http.MethodPatch] 时可以使用此值表示可解析的对象类型；
//
// 如果 method 非空，表示该媒体类型仅对该类型的请求有效。以下几种请求方法有：
//   - 空值，可用于处理所有的请求类型；
//   - [http.MethodPatch] 只能处理由 PATCH 请求提交的数据；
//   - [http.MethodQuery] 只能处理由 QUERY 请求提交的数据；
//
// 对于 patchKind 参数不为零值的情况，还提供了两个简便的方法 [NewPatcherMediaType] 和 [NewMergerMediaType]。
func NewMediaType(codec Codec, name, problem, method string, patchKind PatchKind) *MediaType {
	if codec == nil {
		panic("参数 codec 不能为 nil")
	}
	if name == "" {
		panic("参数 name 不能为空")
	}
	if problem == "" {
		problem = name
	}

	if method != http.MethodPatch && patchKind != PatchKindNone {
		panic("patchKind 只在 method 为 PATCH 时有效")
	}

	return &MediaType{
		Codec:   codec,
		name:    name,
		problem: problem,
		method:  method,
		kind:    patchKind,
	}
}

func NewPatcherMediaType(codec Codec, name, problem string) *MediaType {
	return NewMediaType(codec, name, problem, http.MethodPatch, PatchKindPatcher)
}

func NewMergerMediaType(codec Codec, name, problem string) *MediaType {
	return NewMediaType(codec, name, problem, http.MethodPatch, PatchKindMerger)
}

func (mt *MediaType) Name() string         { return mt.name }
func (mt *MediaType) ProblemName() string  { return mt.problem }
func (mt *MediaType) Method() string       { return mt.method }
func (mt *MediaType) PatchKind() PatchKind { return mt.kind }

// Negotiation 提供与内容协商相关的功能
type Negotiation struct {
	l *locale.Locale // 本地化

	disableCompress      bool
	compressions         []*compression
	acceptEncodingHeader string // 生成 AcceptEncoding 报头内容

	mediaTypes        []*MediaType
	acceptHeader      string // 生成的 Accept 报头内容
	acceptPatchHeader string // 生成的 Accept-Patch 报头内容
	acceptQueryHeader string // 生成的 Accept-Query 报头内容
}

type compression struct {
	compressor     Compressor
	types          []string
	wildcard       bool     // 如果是通配符，则其它配置都将不启作用。
	wildcardSuffix []string // 是模糊类型的，比如 text/*，只有在 types 找不到时，才在此处查找。
}

// NewNegotiation 声明 [Negotiation] 对象
//
// conf 为加载配置文件的对象，[Negotiation.LoadMessages] 使用该值加载本地化的语言文件。
func NewNegotiation(lang language.Tag, conf *config.Config) *Negotiation {
	return &Negotiation{
		compressions: make([]*compression, 0, 10),
		mediaTypes:   make([]*MediaType, 0, 10),
		l:            locale.New(lang, conf),
	}
}

// Language 返回默认的语言标签
func (nego *Negotiation) Language() language.Tag { return nego.l.ID() }

func (nego *Negotiation) Catalog() *catalog.Builder { return nego.l.Catalog() }

func (nego *Negotiation) acceptLanguageHeader() string { return nego.l.AcceptLanguageHeader() }

func (nego *Negotiation) config() *config.Config { return nego.l.Config() }

// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
//
// 根据 [Server.Config] 处理文件格式，如果文件格式不被支持，将无法加载。
func (nego *Negotiation) LoadMessages(glob string, fsys ...fs.FS) error {
	return nego.l.LoadMessages(glob, fsys...)
}

// LocalePrinter 最符合 [Locale.Language] 的 [message.LocalePrinter] 对象
func (nego *Negotiation) LocalePrinter() *message.Printer { return nego.l.Printer() }

// Sprintf 等同于 [Locale.Printer.Sprintf]
func (nego *Negotiation) Sprintf(key string, v ...any) string {
	return nego.LocalePrinter().Sprintf(key, v...)
}

// NewLocalePrinter 声明最符合 tag 的 [message.Printer] 对象
//
// NOTE: 每当更新 [Negotiation.Catalog] 时，可能造成 NewLocalePrinter 相同的入参而返回不同的返回对象的情况。
func (nego *Negotiation) NewLocalePrinter(id language.Tag) *message.Printer {
	return nego.l.NewPrinter(id)
}

// AddCompressor 添加新的压缩算法
//
// t 表示适用的 content-type 类型，可以包含通配符，比如：
//   - application/json
//   - text/*
//   - *
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
func (nego *Negotiation) AddMediaType(mt ...*MediaType) *Negotiation {
	for _, mm := range mt {
		if slices.ContainsFunc(nego.mediaTypes, func(v *MediaType) bool { return v.Name() == mm.Name() }) {
			panic(fmt.Sprintf("存在重复的项 %s", mm.Name()))
		}

		nego.mediaTypes = append(nego.mediaTypes, mm)

		switch m := mm.Method(); m {
		case "":
			if nego.acceptHeader != "" {
				nego.acceptHeader += ", "
			}
			nego.acceptHeader += mm.Name()
		case http.MethodPatch:
			if nego.acceptPatchHeader != "" {
				nego.acceptPatchHeader += ", "
			}
			nego.acceptPatchHeader += mm.Name()
		case "QUERY":
			if nego.acceptQueryHeader != "" {
				nego.acceptQueryHeader += ", "
			}
			nego.acceptQueryHeader += mm.Name()
		}
	}

	return nego
}

// AddCodec 添加编码
func (nego *Negotiation) AddCodec(codec Codec, name, problem, method string, kind PatchKind) *Negotiation {
	return nego.AddMediaType(NewMediaType(codec, name, problem, method, kind))
}

func unableParse(v string) error { return localeutil.Error("unable parse %s", v) }

// 根据 Accept-Language 报头获取最适合的语言
func (nego *Negotiation) acceptLanguage(h string) language.Tag {
	if h == "" {
		return nego.Language()
	}
	tag, _ := language.MatchStrings(nego.Catalog().Matcher(), h)
	return tag
}

// 根据 Accept-Encoding 报头选择是适合的压缩方法
//
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
	// *;q=0 表示匹配所有未列出的编码

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
			if !mediaTypeIsDenied(c.compressor.Name(), deny) {
				return c.compressor, nil
			}
		}
	}

	if slices.Contains(deny, header.Identity) { // 明确将 identity 设置为 0
		return nil, localeutil.Error("client disabled identity")
	}
	return nil, nil
}

// 从 Accept 报头解析出所需要的解码函数
//
// 如果 h 为空，那么将被设置为默认值 ＊/*；
func (nego *Negotiation) accept(h string) (*MediaType, error) {
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
			name, mn := item.Value, mt.Name()
			if name == mn || matchPrefix(mn, name) {
				return mt, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 */*，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, mt := range nego.mediaTypes {
			if !mediaTypeIsDenied(mt.Name(), deny) {
				return mt, nil
			}
		}
	}

	return nil, nil
}

// name 是否在 deny 中
func mediaTypeIsDenied(name string, deny []string) bool {
	for _, d := range deny {
		if d == "*/*" || d == name || matchPrefix(name, d) {
			return true
		}
	}
	return false
}

// match 为 XX/* 且 v 符合 match 的规则
func matchPrefix(v, match string) bool {
	return strings.HasSuffix(match, "/*") && strings.HasPrefix(v, match[:len(match)-1])
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

// 根据 Content-Encoding 报头对 r 进行包装
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

// 从 Content-Type 报头中获取解码和字符集函数
//
// h 表示 Content-Type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
// method 如果为空表示只从 [MediaType.Method] 为空的类型中查找，
// 如果不为空，除了前者还会从 [MediaType.Method] 与 method 相同的类型中查找；
func (nego *Negotiation) contentType(h, method string) (*MediaType, encoding.Encoding, error) {
	mt, charset := qheader.ParseWithParam(h, "charset")

	item, found := sliceutil.At(nego.mediaTypes, func(i *MediaType, _ int) bool {
		return (i.Name() == mt) && (method == "" || i.Method() == "" || i.Method() == method)
	})
	if !found {
		return nil, nil, NewError("not found serialization function for %s", mt)
	}

	if charset == "" || charset == header.UTF8 {
		return item, nil, nil
	}
	c, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return item, c, nil
}

// CanCompress 当前是否拥有压缩功能
func (nego *Negotiation) CanCompress() bool { return !nego.disableCompress }

// SetCompress 设置压缩功能
//
// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
func (nego *Negotiation) SetCompress(enable bool) { nego.disableCompress = !enable }

// MediaTypes 返回支持的所有媒体类型
func (nego *Negotiation) MediaTypes() iter.Seq[*MediaType] {
	return func(yield func(*MediaType) bool) {
		for _, v := range nego.mediaTypes {
			if !yield(v) {
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
