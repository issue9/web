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

	"github.com/issue9/mux/v10/header"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/internal/qheader"
)

// Codec 编解码工具
//
// 包含了压缩方法和媒体类型的处理
type Codec struct {
	disableCompress      bool
	compressions         []*compression
	acceptEncodingHeader string // 生成 AcceptEncoding 报头内容

	mediaTypes         []*mediaType
	clientAcceptHeader string // 生成客户端的 Accept 报头内容
	serverAcceptHeader string // 生成服务端的 Accept 报头内容
}

type mediaType struct {
	// Mimetype 的名称
	//
	// 比如：application/json
	Name string

	// 对应的错误状态下的 mimetype 值
	//
	// 比如：application/problem+json。
	// 可以为空，表示与 Type 相同。
	Problem string

	// 生成编码方法
	Marshal MarshalFunc

	// 解码方法
	Unmarshal UnmarshalFunc

	requestAccept, responseAccept bool
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
		mediaTypes:   make([]*mediaType, 0, 10),
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

// AddMimetype 添加对媒体类型的编解码函数
//
// name 媒体类型的名称，比如 application/json；
// problem 媒体类型非正常状态下的子类型，比如 application/problem+json；
// requestAccept 是否出现在客户端请求的 accept 报头中；
// responseAccept 是否出现在服务端返回的 accept 报头中；
func (c *Codec) AddMimetype(name string, m MarshalFunc, u UnmarshalFunc, problem string, requestAccept, responseAccept bool) *Codec {
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

	// 检测复复值
	if slices.IndexFunc(c.mediaTypes, func(v *mediaType) bool { return v.Name == name }) >= 0 {
		panic(fmt.Sprintf("存在重复的项 %s", name))
	}

	c.mediaTypes = append(c.mediaTypes, &mediaType{
		Name:           name,
		Marshal:        m,
		Unmarshal:      u,
		Problem:        problem,
		requestAccept:  requestAccept,
		responseAccept: responseAccept,
	})

	if requestAccept {
		if c.clientAcceptHeader != "" {
			c.clientAcceptHeader += ", "
		}
		c.clientAcceptHeader += name
	}

	if responseAccept {
		if c.serverAcceptHeader != "" {
			c.serverAcceptHeader += ", "
		}
		c.serverAcceptHeader += name
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
	return nil, NewLocaleError("not found compress for %s", name)
}

// 根据客户端的 Accept-Encoding 报头选择是适合的压缩方法
//
// 如果返回的 c 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
func (e *Codec) acceptEncoding(contentType, h string) (c Compressor, notAcceptable bool) {
	if len(e.compressions) == 0 {
		return
	}

	accepts := qheader.ParseQHeader(h, "*")
	defer qheader.PutQHeader(accepts)
	if len(accepts) == 0 {
		return
	}

	indexes := e.getMatchCompresses(contentType)
	if len(indexes) == 0 {
		return
	}

	if last := accepts[len(accepts)-1]; last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return nil, true
		}

		for _, index := range indexes {
			curr := e.compressions[index]
			if slices.IndexFunc(accepts, func(i *qheader.Item) bool { return i.Value == curr.compressor.Name() }) < 0 {
				return curr.compressor, false
			}
		}
		return
	}

	var identity *qheader.Item
	for _, accept := range accepts {
		if accept.Err != nil {
			// NOTE: 对于客户端的错误，不记录于日志中。
			continue
		}

		if accept.Value == header.Identity { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, index := range indexes {
			if curr := e.compressions[index]; curr.compressor.Name() == accept.Value {
				return curr.compressor, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		c := e.compressions[indexes[0]]
		return c.compressor, false
	}

	return // 没有匹配，表示不需要进行压缩
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

func (m *mediaType) name(problem bool) string {
	if problem {
		return m.Problem
	}
	return m.Name
}

// 从请求端提交的 Content-Type 报头中获取解码和字符集函数
//
// h 表示 Content-Type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (e *Codec) contentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mimetype, charset := qheader.ParseWithParam(h, "charset")

	item := e.searchFunc(func(s string) bool { return s == mimetype })
	if item == nil {
		return nil, nil, NewLocaleError("not found serialization function for %s", mimetype)
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
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//
//	application/json;q=0.9,*/*;q=1
//
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
func (c *Codec) accept(h string) *mediaType {
	if h == "" {
		if item := c.findMarshal("*/*"); item != nil {
			return item
		}
		return nil
	}

	items := qheader.ParseQHeader(h, "*/*")
	defer qheader.PutQHeader(items)
	for _, item := range items {
		if i := c.findMarshal(item.Value); i != nil {
			return i
		}
	}

	return nil
}

func (c *Codec) findMarshal(name string) *mediaType {
	switch {
	case len(c.mediaTypes) == 0:
		return nil
	case name == "" || name == "*/*":
		return c.searchFunc(func(s string) bool { return true }) // 第一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		return c.searchFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		return c.searchFunc(func(s string) bool { return s == name })
	}
}

func (c *Codec) searchFunc(match func(string) bool) *mediaType {
	item, _ := sliceutil.At(c.mediaTypes, func(i *mediaType, _ int) bool { return match(i.Name) || match(i.Problem) })
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

// Mimetypes 返回支持的所有媒体类型
//
// 键名为 mimetype，键值为 problem 状态下的 mimetype。
func (c *Codec) Mimetypes() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, v := range c.mediaTypes {
			if !yield(v.Name, v.Problem) {
				break
			}
		}
	}
}
