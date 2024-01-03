// SPDX-License-Identifier: MIT

package server

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
	"github.com/issue9/web/codec/compressor"
	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/codec/mimetype/xml"
	"github.com/issue9/web/locales"
)

// Compression 有关压缩的设置项
type Compression struct {
	// Compressor 压缩算法
	Compressor compressor.Compressor

	// Types 该压缩对象允许使用的为 content-type 类型
	//
	// 如果是 * 或是空值表示适用所有类型。
	Types []string
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

func buildCodec(ms []*Mimetype, cs []*Compression) (*web.Codec, *web.FieldError) {
	if len(ms) == 0 {
		ms = JSONMimetypes()
	}

	// 检测是否存在同名的项
	indexes := sliceutil.Dup(ms, func(e1, e2 *Mimetype) bool { return e1.Name == e2.Name })
	if len(indexes) > 0 {
		return nil, web.NewFieldError("Mimetypes["+strconv.Itoa(indexes[0])+"].Name", locales.DuplicateValue)
	}

	c := web.NewCodec()

	for i, s := range ms {
		if s.Name == "" {
			return nil, web.NewFieldError("Mimetypes["+strconv.Itoa(i)+"].Name", locales.CanNotBeEmpty)
		}

		if s.Marshal == nil {
			return nil, web.NewFieldError("Mimetypes["+strconv.Itoa(i)+"].Marshal", locales.CanNotBeEmpty)
		}

		if s.Unmarshal == nil {
			return nil, web.NewFieldError("Mimetypes["+strconv.Itoa(i)+"].Unmarshal", locales.CanNotBeEmpty)
		}

		c.AddMimetype(s.Name, s.Marshal, s.Unmarshal, s.Problem)
	}

	for i, s := range cs {
		if s.Compressor == nil {
			return nil, web.NewFieldError("Compressions["+strconv.Itoa(i)+"].Compressor", locales.CanNotBeEmpty)
		}
		c.AddCompressor(s.Compressor, s.Types...)
	}

	return c, nil
}

// DefaultCompressions 提供当前框架内置的所有压缩算法
//
// contentType 指定所有算法应用的媒体类型，为空则表示对所有的内容都进行压缩。
func DefaultCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Compressor: compressor.NewGzip(gzip.DefaultCompression), Types: contentType},
		{Compressor: compressor.NewDeflate(flate.DefaultCompression, nil), Types: contentType},
		{Compressor: compressor.NewLZW(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotli(brotli.WriterOptions{}), Types: contentType},
		{Compressor: compressor.NewZstd(), Types: contentType},
	}
}

// BestSpeedCompressions 提供当前框架内置的所有压缩算法
//
// 如果有性能参数，则选择最快速度作为初始化条件。
func BestSpeedCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Compressor: compressor.NewGzip(gzip.BestSpeed), Types: contentType},
		{Compressor: compressor.NewDeflate(flate.BestSpeed, nil), Types: contentType},
		{Compressor: compressor.NewLZW(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.BestSpeed}), Types: contentType},
		{Compressor: compressor.NewZstd(), Types: contentType},
	}
}

// BestCompressionCompressions 提供当前框架内置的所有压缩算法
//
// 如果有性能参数，则选择最快压缩比作为初始化条件。
func BestCompressionCompressions(contentType ...string) []*Compression {
	return []*Compression{
		{Compressor: compressor.NewGzip(gzip.BestCompression), Types: contentType},
		{Compressor: compressor.NewDeflate(flate.BestCompression, nil), Types: contentType},
		{Compressor: compressor.NewLZW(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.BestCompression}), Types: contentType},
		{Compressor: compressor.NewZstd(), Types: contentType},
	}
}

// APIMimetypes 返回以 XML 和 JSON 作为数据交换格式的配置项
func APIMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, Marshal: json.Marshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: xml.Mimetype, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
	}
}

// XMLMimetypes 返回以 XML 作为数据交换格式的配置项
func XMLMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: xml.Mimetype, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
	}
}

// JSONMimetypes 返回以 JSON 作为数据交换格式的配置项
func JSONMimetypes() []*Mimetype {
	return []*Mimetype{
		{Name: json.Mimetype, Marshal: json.Marshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
	}
}
