// SPDX-License-Identifier: MIT

// Package codec 编码解码工具
//
// 包含了压缩方法和媒体类型的处理，实现了 [web.Codec] 接口及相关内容。
package codec

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/issue9/config"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
	"github.com/issue9/web/codec/compressor"
	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/codec/mimetype/xml"
	"github.com/issue9/web/locales"
)

type codec struct {
	compressions         []*Compression
	acceptEncodingHeader string
	disableCompress      bool

	types        []*mimetype
	acceptHeader string
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
	Marshal web.MarshalFunc

	// 解码方法
	Unmarshal web.UnmarshalFunc
}

// Compression 有关压缩的设置项
type Compression struct {
	// Compressor 压缩算法
	Compressor compressor.Compressor

	// Types 该压缩对象允许使用的为 content-type 类型
	//
	// 如果是 * 表示适用所有类型。
	Types []string

	// 如果是通配符，则其它配置都将不启作用。
	wildcard bool

	// Types 是具体值的，比如 text/xml
	// wildcardSuffix 是模糊类型的，比如 text/*，只有在 Types 找不到时，才在此处查找。

	wildcardSuffix []string
}

func (m *Mimetype) sanitize() *web.FieldError {
	if m.Name == "" {
		return web.NewFieldError("Name", locales.CanNotBeEmpty)
	}

	if m.Problem == "" {
		m.Problem = m.Name
	}

	return nil
}

func (m *Compression) sanitize() *web.FieldError {
	if m.Compressor == nil {
		return web.NewFieldError("Compressor", locales.CanNotBeEmpty)
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

// APIMimetypes 返回以 XML 和 JSON 作为数据交换格式的配置项
func APIMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, Marshal: json.Marshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: xml.Mimetype, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}
}

// XMLMimetypes 返回以 XML 作为数据交换格式的配置项
func XMLMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: xml.Mimetype, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}
}

// JSONMimetypes 返回以 JSON 作为数据交换格式的配置项
func JSONMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, Marshal: json.Marshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}
}

// DefaultCompressions 提供当前框架内置的所有压缩算法
//
// contentType 指定所有算法应用的媒体类型，为空则表示对所有的内容都进行压缩。
func DefaultCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Compressor: compressor.NewGzipCompressor(gzip.DefaultCompression), Types: contentType},
		{Compressor: compressor.NewDeflateCompressor(flate.DefaultCompression, nil), Types: contentType},
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: contentType},
		{Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}

// BestSpeedCompressions 提供当前框架内置的所有压缩算法
//
// 如果有性能参数，则选择最快速度作为初始化条件。
func BestSpeedCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Compressor: compressor.NewGzipCompressor(gzip.BestSpeed), Types: contentType},
		{Compressor: compressor.NewDeflateCompressor(flate.BestSpeed, nil), Types: contentType},
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestSpeed}), Types: contentType},
		{Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}

// BestCompressionCompressions 提供当前框架内置的所有压缩算法
//
// 如果有性能参数，则选择最快压缩比作为初始化条件。
func BestCompressionCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Compressor: compressor.NewGzipCompressor(gzip.BestCompression), Types: contentType},
		{Compressor: compressor.NewDeflateCompressor(flate.BestCompression, nil), Types: contentType},
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestCompression}), Types: contentType},
		{Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}

// New 声明 [web.Codec] 对象
//
// csName 和 msName 分别表示 cs 和 ms 在出错时在返回对象中的字段名称。
func New(msName, csName string, ms []*Mimetype, cs []*Compression) (web.Codec, *web.FieldError) {
	c := &codec{
		compressions: make([]*Compression, 0, len(cs)),
		types:        make([]*mimetype, 0, len(ms)),
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
