// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"io"
	"iter"
	"slices"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

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

// Codec 编解码工具
//
// 包含了压缩方法和媒体类型的处理
type Codec struct {
	disableCompress      bool
	compressions         []*compression
	acceptEncodingHeader string // 生成 AcceptEncoding 报头内容

	mediaTypes         []*MediaType
	clientAcceptHeader string // 生成客户端的 Accept 报头内容
	serverAcceptHeader string // 生成服务端的 Accept 报头内容
}

type MediaType struct {
	// 媒体类型的名称
	//
	// 比如：application/json
	normalName string

	// 对应的错误状态下的值
	//
	// 比如：application/problem+json。如果为空则采用 normalName。
	problemName string

	// 生成编码方法
	marshal MarshalFunc

	// 解码方法
	unmarshal UnmarshalFunc
}

type compression struct {
	compressor Compressor

	types []string

	// 如果是通配符，则其它配置都将不启作用。
	wildcard bool

	// 是模糊类型的，比如 text/*，只有在 Types 找不到时，才在此处查找。
	wildcardSuffix []string
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

// NewCodec 声明 [Codec] 对象
func NewCodec() *Codec {
	return &Codec{
		compressions: make([]*compression, 0, 10),
		mediaTypes:   make([]*MediaType, 0, 10),
	}
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
func (e *Codec) AddCompressor(c Compressor, t ...string) *Codec {
	e.compressions = append(e.compressions, buildCompression(c, t))

	names := make([]string, 0, len(e.compressions))
	for _, item := range e.compressions {
		names = append(names, item.compressor.Name())
	}
	names = sliceutil.Unique(names, func(i, j string) bool { return i == j })
	e.acceptEncodingHeader = strings.Join(names, ",")

	return e
}

// NewMediaType 声明一个媒体类型
//
// name 媒体类型的名称，比如 application/json；
// problem 媒体类型非正常状态下的子类型，比如 application/problem+json；
func NewMediaType(name, problem string, m MarshalFunc, u UnmarshalFunc) *MediaType {
	if problem == "" {
		problem = name
	}

	if name == "" {
		panic("参数 name 不能为空")
	}

	if m == nil {
		panic("参数 m 不能为空")
	}

	if u == nil {
		panic("参数 u 不能为空")
	}

	return &MediaType{
		normalName:  name,
		problemName: problem,
		marshal:     m,
		unmarshal:   u,
	}
}

// AddMediaType 添加媒体类型
//
// requestAccept 是否出现在客户端请求的 accept 报头中；
// responseAccept 是否出现在服务端返回的 accept 报头中；
func (c *Codec) AddMediaType(mt *MediaType, requestAccept, responseAccept bool) *Codec {
	if mt == nil {
		panic("参数 mt 不能为空")
	}

	// 检测复复值
	if slices.IndexFunc(c.mediaTypes, func(v *MediaType) bool { return v.normalName == mt.normalName }) >= 0 {
		panic(fmt.Sprintf("存在重复的项 %s", mt.normalName))
	}

	c.mediaTypes = append(c.mediaTypes, mt)

	if requestAccept {
		if c.clientAcceptHeader != "" {
			c.clientAcceptHeader += ", "
		}
		c.clientAcceptHeader += mt.normalName
	}

	if responseAccept {
		if c.serverAcceptHeader != "" {
			c.serverAcceptHeader += ", "
		}
		c.serverAcceptHeader += mt.normalName
	}

	return c
}

// 根据客户端的 Content-Encoding 报头对 r 进行包装
//
// name 编码名称，即 Content-Encoding 报头内容；
// r 为未解码的内容；
func (c *Codec) contentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if name == "" {
		return io.NopCloser(r), nil
	}

	if c, f := sliceutil.At(c.compressions, func(item *compression, _ int) bool { return item.compressor.Name() == name }); f {
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
func (e *Codec) acceptEncoding(contentType, h string) (c Compressor, err error) {
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
		if slices.Contains(deny, header.Identity) { // 且服务端明确必须要压缩
			return nil, localeutil.Error("client set \"identity;q=0\"")
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
		return nil, localeutil.Error("client set \"identity;q=0\"")
	}
	return nil, nil
}

func (c *Codec) getMatchCompresses(contentType string) []int {
	indexes := make([]int, 0, len(c.compressions))

LOOP:
	for index, c := range c.compressions {
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
func (e *Codec) contentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mt, charset := qheader.ParseWithParam(h, "charset")

	item := e.searchFunc(func(s string) bool { return s == mt })
	if item == nil {
		return nil, nil, NewError("not found serialization function for %s", mt)
	}

	if charset == "" || charset == header.UTF8 {
		return item.unmarshal, nil, nil
	}
	c, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return item.unmarshal, c, nil
}

// 从请求端提交的 Accept 报头解析出所需要的解码函数
//
// 如果 h 为空，那么将被设置为默认值 ＊；
func (c *Codec) accept(h string) (*MediaType, error) {
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

		for _, mt := range c.mediaTypes {
			if match(item.Value, mt) {
				return mt, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 */*，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, mt := range c.mediaTypes {
			if !denied(mt.normalName, deny) {
				return mt, nil
			}
		}
	}

	return nil, nil
}

func match(name string, mt *MediaType) bool {
	switch {
	case name == mt.normalName:
		return true
	case strings.HasSuffix(name, "/*") && strings.HasPrefix(mt.normalName, name[:len(name)-1]):
		return true
	}
	return false
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

func (c *Codec) searchFunc(match func(string) bool) *MediaType {
	item, _ := sliceutil.At(c.mediaTypes, func(i *MediaType, _ int) bool { return match(i.normalName) || match(i.problemName) })
	return item
}

// CanCompress 当前是否拥有压缩功能
func (c *Codec) CanCompress() bool { return !c.disableCompress }

// SetCompress 设置压缩功能
//
// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
//
// NOTE: 仅对输出内容启作用，读取内容始终是按照提交的 Content-Encoding 指定算法进行解析。
func (c *Codec) SetCompress(enable bool) { c.disableCompress = !enable }

// MediaTypes 返回支持的所有媒体类型
//
// 键名和键值分别为正常和 problem 下的名称。
func (c *Codec) MediaTypes() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, v := range c.mediaTypes {
			if !yield(v.normalName, v.problemName) {
				break
			}
		}
	}
}

// Name 获取媒体类型的名称
func (m *MediaType) Name(problem bool) string {
	if problem {
		return m.problemName
	}
	return m.normalName
}

// Marshal 对内容进行编码
func (m *MediaType) Marshal(ctx *Context, v any) ([]byte, error) { return m.marshal(ctx, v) }

// Marshal 解码内容到 v
func (m *MediaType) Unmarshal(r io.Reader, v any) error { return m.unmarshal(r, v) }
