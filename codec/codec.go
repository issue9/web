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
	compressions         []*compression
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
	MarshalBuilder web.BuildMarshalFunc

	// 解码方法
	Unmarshal web.UnmarshalFunc
}

// Compression 有关压缩的设置项
type Compression struct {
	// Name 压缩方法的名称
	//
	// 可以重名，比如 gzip，可以配置参数不同的对象。
	Name string

	// Compressor 压缩算法
	Compressor compressor.Compressor

	// Types 该压缩对象允许使用的为 content-type 类型
	//
	// 如果是 * 表示适用所有类型。
	Types []string
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
	if m.Name == "" {
		return web.NewFieldError("Name", locales.CanNotBeEmpty)
	}

	if len(m.Types) == 0 {
		m.Types = []string{"*"}
	}

	return nil
}

// APIMimetypes 返回以 XML 和 JSON 作为数据交换格式的配置项
func APIMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: xml.Mimetype, MarshalBuilder: xml.BuildMarshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", MarshalBuilder: nil, Unmarshal: nil},
	}
}

// XMLMimetypes 返回以 XML 作为数据交换格式的配置项
func XMLMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: xml.Mimetype, MarshalBuilder: xml.BuildMarshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", MarshalBuilder: nil, Unmarshal: nil},
	}
}

// JSONMimetypes 返回以 JSON 作为数据交换格式的配置项
func JSONMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: "nil", MarshalBuilder: nil, Unmarshal: nil},
	}
}

// DefaultCompressions 提供当前框架内置的所有压缩算法
//
// contentType 指定所有算法应用的媒体类型，为空则表示对所有的内容都进行压缩。
func DefaultCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Name: "gzip", Compressor: compressor.NewGzipCompressor(gzip.DefaultCompression), Types: contentType},
		{Name: "deflate", Compressor: compressor.NewDeflateCompressor(flate.DefaultCompression, nil), Types: contentType},
		{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Name: "br", Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: contentType},
		{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}

// BestSpeedCompressions 提供当前框架内置的所有压缩算法
//
// 如果有性能参数，则选择最快速度作为初始化条件。
func BestSpeedCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Name: "gzip", Compressor: compressor.NewGzipCompressor(gzip.BestSpeed), Types: contentType},
		{Name: "deflate", Compressor: compressor.NewDeflateCompressor(flate.BestSpeed, nil), Types: contentType},
		{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Name: "br", Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestSpeed}), Types: contentType},
		{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}

// BestCompressionCompressions 提供当前框架内置的所有压缩算法
//
// 如果有性能参数，则选择最快压缩比作为初始化条件。
func BestCompressionCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Name: "gzip", Compressor: compressor.NewGzipCompressor(gzip.BestCompression), Types: contentType},
		{Name: "deflate", Compressor: compressor.NewDeflateCompressor(flate.BestCompression, nil), Types: contentType},
		{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Name: "br", Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestCompression}), Types: contentType},
		{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}

// New 声明 [web.Codec] 对象
//
// csName 和 msName 分别表示 cs 和 ms 在出错时在返回对象中的字段名称。
func New(msName, csName string, ms []*Mimetype, cs []*Compression) (web.Codec, *web.FieldError) {
	c := &codec{
		compressions: make([]*compression, 0, len(cs)),
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
