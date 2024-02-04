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
	"github.com/issue9/web/compressor"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/xml"
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

type compressConfig struct {
	// Type content-type 的值
	//
	// 可以带通配符，比如 text/* 表示所有 text/ 开头的 content-type 都采用此压缩方法。
	Types []string `json:"types" xml:"type" yaml:"types"`

	// IDs 压缩方法的 ID 列表
	//
	// 这些 ID 值必须是由 [RegisterCompress] 注册的，否则无效，默认情况下支持以下类型：
	//  - deflate-default
	//  - deflate-best-compression
	//  - deflate-best-speed
	//  - gzip-default
	//  - gzip-best-compression
	//  - gzip-best-speed
	//  - compress-lsb-8
	//  - compress-msb-8
	//  - br-default
	//  - br-best-compression
	//  - br-best-speed
	//  - zstd-default
	ID string `json:"id" xml:"id,attr" yaml:"id"`
}

type mimetypeConfig struct {
	// 编码名称
	//
	// 比如 application/xml 等
	Type string `json:"type" yaml:"type" xml:"type,attr"`

	// 返回错误代码是的 mimetype
	//
	// 比如正常情况下如果是 application/json，那么此值可以是 application/problem+json。
	// 如果为空，表示与 Type 相同。
	Problem string `json:"problem,omitempty" yaml:"problem,omitempty" xml:"problem,attr,omitempty"`

	// 实际采用的解码方法
	//
	// 由 [RegisterMimetype] 注册而来。默认可用为：
	//
	//  - xml
	//  - json
	//  - form
	//  - html
	//  - gob
	//  - nop  没有具体实现的方法，对于上传等需要自行处理的情况可以指定此值。
	Target string `json:"target" yaml:"target" xml:"target,attr"`
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

func (conf *configOf[T]) sanitizeCompresses() *web.FieldError {
	conf.compressors = make([]*Compression, 0, len(conf.Compressors))
	for index, e := range conf.Compressors {
		enc, found := compressorFactory.get(e.ID)
		if !found {
			field := "compresses[" + strconv.Itoa(index) + "].id"
			return web.NewFieldError(field, locales.ErrNotFound(e.ID))
		}

		conf.compressors = append(conf.compressors, &Compression{
			Compressor: enc,
			Types:      e.Types,
		})
	}
	return nil
}

func (conf *configOf[T]) sanitizeMimetypes() *web.FieldError {
	indexes := sliceutil.Dup(conf.Mimetypes, func(i, j *mimetypeConfig) bool { return i.Type == j.Type })
	if len(indexes) > 0 {
		value := conf.Mimetypes[indexes[1]].Type
		err := web.NewFieldError("mimetypes["+strconv.Itoa(indexes[1])+"].target", locales.DuplicateValue)
		err.Value = value
		return err
	}

	ms := make([]*Mimetype, 0, len(conf.Mimetypes))
	for index, item := range conf.Mimetypes {
		m, found := mimetypesFactory.get(item.Target)
		if !found {
			return web.NewFieldError("mimetypes["+strconv.Itoa(index)+"].target", locales.ErrNotFound(item.Target))
		}

		ms = append(ms, &Mimetype{
			Marshal:   m.marshal,
			Unmarshal: m.unmarshal,
			Name:      item.Type,
			Problem:   item.Problem,
		})
	}
	conf.mimetypes = ms

	return nil
}
