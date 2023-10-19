// SPDX-License-Identifier: MIT

// Package codec 编码解码工具
//
// 包含了压缩方法和媒体类型的处理。
package codec

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"

	"github.com/andybalholm/brotli"

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

func (m *Mimetype) SanitizeConfig() *web.FieldError {
	if m.Name == "" {
		return web.NewFieldError("Name", locales.CanNotBeEmpty)
	}

	if m.Problem == "" {
		m.Problem = m.Name
	}

	return nil
}

func (m *Compression) SanitizeConfig() *web.FieldError {
	if m.Name == "" {
		return web.NewFieldError("Name", locales.CanNotBeEmpty)
	}

	if len(m.Types) == 0 {
		m.Types = []string{"*"}
	}

	return nil
}

func APIMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: xml.Mimetype, MarshalBuilder: xml.BuildMarshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", MarshalBuilder: nil, Unmarshal: nil},
	}
}

func XMLMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: xml.Mimetype, MarshalBuilder: xml.BuildMarshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", MarshalBuilder: nil, Unmarshal: nil},
	}
}

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
// 用户需要自行调用 ms 和 cs 的 [config.Sanitizer] 接口对数据合规性作检测。
func New(ms []*Mimetype, cs []*Compression) web.Codec {
	c := &codec{
		compressions: make([]*compression, 0, len(cs)),
		types:        make([]*mimetype, 0, len(ms)),
	}

	for _, m := range ms {
		c.addMimetype(m)
	}

	for _, cc := range cs {
		c.addCompression(cc)
	}

	return c
}
