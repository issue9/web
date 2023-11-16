// SPDX-License-Identifier: MIT

package web

import (
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/logs"
)

// Codec 编码解码工具
//
// 包含了压缩方法和媒体类型的处理
type Codec struct {
	compressions         []*Compression
	acceptEncodingHeader string // 生成 AcceptEncoding 报头内容

	types        []*Mimetype
	acceptHeader string // 生成 Accept 报头内容
}

// Compressor 压缩算法的接口
type Compressor interface {
	// Name 算法的名称
	Name() string

	// NewDecoder 将 r 包装成为当前压缩算法的解码器
	NewDecoder(r io.Reader) (io.ReadCloser, error)

	// NewEncoder 将 w 包装成当前压缩算法的编码器
	NewEncoder(w io.Writer) (io.WriteCloser, error)
}

// Mimetype 有关 mimetype 的设置项
type Mimetype struct {
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
}

// Compression 有关压缩的设置项
type Compression struct {
	// Compressor 压缩算法
	Compressor Compressor

	// Types 该压缩对象允许使用的为 content-type 类型
	//
	// 如果是 * 或是空值表示适用所有类型。
	Types []string

	// 如果是通配符，则其它配置都将不启作用。
	wildcard bool

	// Types 是具体值的，比如 text/xml
	// wildcardSuffix 是模糊类型的，比如 text/*，只有在 Types 找不到时，才在此处查找。

	wildcardSuffix []string
}

func (m *Mimetype) sanitize() *FieldError {
	if m.Name == "" {
		return NewFieldError("Name", locales.CanNotBeEmpty)
	}

	if m.Problem == "" {
		m.Problem = m.Name
	}

	return nil
}

func (m *Compression) sanitize() *FieldError {
	if m.Compressor == nil {
		return NewFieldError("Compressor", locales.CanNotBeEmpty)
	}

	if len(m.Types) == 0 {
		m.wildcard = true
		return nil
	}

	types := make([]string, 0, len(m.Types))
	suffix := make([]string, 0, len(m.Types))
	for _, c := range m.Types {
		if c == "" {
			continue
		}

		if c == "*" {
			m.Types = nil
			m.wildcardSuffix = nil
			m.wildcard = true
			return nil
		}

		if c[len(c)-1] == '*' {
			suffix = append(suffix, c[:len(c)-1])
		} else {
			types = append(types, c)
		}
	}

	m.Types = types
	m.wildcardSuffix = suffix

	return nil
}

// NewCodec 声明 [Codec] 对象
//
// csName 和 msName 分别表示 cs 和 ms 在出错时在返回对象中的字段名称。
func NewCodec(msName, csName string, ms []*Mimetype, cs []*Compression) (*Codec, *FieldError) {
	c := &Codec{
		compressions: make([]*Compression, 0, len(cs)),
		types:        make([]*Mimetype, 0, len(ms)),
	}

	for i, s := range ms {
		if err := s.sanitize(); err != nil {
			err.AddFieldParent(msName + "[" + strconv.Itoa(i) + "]")
			return nil, err
		}
	}
	indexes := sliceutil.Dup(ms, func(e1, e2 *Mimetype) bool { return e1.Name == e2.Name })
	if len(indexes) > 0 {
		return nil, config.NewFieldError(msName+"["+strconv.Itoa(indexes[0])+"].Name", locales.DuplicateValue)
	}

	for i, s := range cs {
		if err := s.sanitize(); err != nil {
			err.AddFieldParent(csName + "[" + strconv.Itoa(i) + "]")
			return nil, err
		}
	}

	for _, m := range ms {
		c.addMimetype(m)
	}

	for _, cc := range cs {
		c.addCompression(cc)
	}

	return c, nil
}

func (e *Codec) addCompression(c *Compression) {
	cc := *c // 复制，防止通过配置项修改内容。
	e.compressions = append(e.compressions, &cc)

	names := make([]string, 0, len(e.compressions))
	for _, item := range e.compressions {
		names = append(names, item.Compressor.Name())
	}
	names = sliceutil.Unique(names, func(i, j string) bool { return i == j })
	e.acceptEncodingHeader = strings.Join(names, ",")
}

// 根据客户端的 Content-Encoding 报头对 r 进行包装
//
// name 编码名称，即 Content-Encoding 报头内容；
// r 为未解码的内容；
func (e *Codec) contentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if name == "" {
		return io.NopCloser(r), nil
	}

	if c, f := sliceutil.At(e.compressions, func(item *Compression, _ int) bool { return item.Compressor.Name() == name }); f {
		return c.Compressor.NewDecoder(r)
	}
	return nil, localeutil.Error("not found compress for %s", name)
}

// 根据客户端的 Accept-Encoding 报头选择是适合的压缩方法
//
// 如果返回的 c 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
// l 表示解析报头过程中的错误信息，可以为空，表示不输出信息；
func (e *Codec) acceptEncoding(contentType, h string, l *logs.Logger) (c Compressor, notAcceptable bool) {
	if len(e.compressions) == 0 {
		return
	}

	accepts := header.ParseQHeader(h, "*")
	defer header.PutQHeader(&accepts)
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
			if !sliceutil.Exists(accepts, func(i *header.Item, _ int) bool { return i.Value == curr.Compressor.Name() }) {
				return curr.Compressor, false
			}
		}
		return
	}

	var identity *header.Item
	for _, accept := range accepts {
		if accept.Err != nil && l != nil {
			l.Error(accept.Err)
			continue
		}

		if accept.Value == header.Identity { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, index := range indexes {
			if curr := e.compressions[index]; curr.Compressor.Name() == accept.Value {
				return curr.Compressor, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		c := e.compressions[indexes[0]]
		return c.Compressor, false
	}

	return // 没有匹配，表示不需要进行压缩
}

func (e *Codec) getMatchCompresses(contentType string) []int {
	indexes := make([]int, 0, len(e.compressions))

LOOP:
	for index, c := range e.compressions {
		if c.wildcard {
			indexes = append(indexes, index)
			continue
		}

		for _, s := range c.Types {
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

func (m *Mimetype) name(problem bool) string {
	if problem {
		return m.Problem
	}
	return m.Name
}

func (e *Codec) addMimetype(m *Mimetype) {
	t := *m // 防止通过配置项修改内容
	e.types = append(e.types, &t)

	names := make([]string, 0, len(e.types))
	for _, item := range e.types {
		if !reflect.ValueOf(item.Unmarshal).IsZero() {
			names = append(names, item.Name)
		}
	}
	e.acceptHeader = strings.Join(names, ",")
}

// 从请求端提交的 Content-Type 报头中获取解码和字符集函数
//
// h 表示 Content-Type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (e *Codec) contentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mimetype, charset := header.ParseWithParam(h, "charset")

	item := e.searchFunc(func(s string) bool { return s == mimetype })
	if item == nil {
		return nil, nil, localeutil.Error("not found serialization function for %s", mimetype)
	}
	f := item.Unmarshal

	if charset == "" || charset == header.UTF8Name {
		return f, nil, nil
	}
	c, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, c, nil
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
func (e *Codec) accept(h string) *Mimetype {
	if h == "" {
		if item := e.findMarshal("*/*"); item != nil {
			return item
		}
		return nil
	}

	items := header.ParseQHeader(h, "*/*")
	defer header.PutQHeader(&items)
	for _, item := range items {
		if i := e.findMarshal(item.Value); i != nil {
			return i
		}
	}

	return nil
}

func (e *Codec) findMarshal(name string) *Mimetype {
	switch {
	case len(e.types) == 0:
		return nil
	case name == "" || name == "*/*":
		return e.searchFunc(func(s string) bool { return true }) // 第一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		return e.searchFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		return e.searchFunc(func(s string) bool { return s == name })
	}
}

func (e *Codec) searchFunc(match func(string) bool) *Mimetype {
	item, _ := sliceutil.At(e.types, func(i *Mimetype, _ int) bool { return match(i.Name) || match(i.Problem) })
	return item
}
