// SPDX-License-Identifier: MIT

// Package codec 编码解码工具
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
)

// APIMimetypes 返回以 XML 和 JSON 作为数据交换格式的配置项
func APIMimetypes() []*web.Mimetype {
	return []*web.Mimetype{
		{Name: json.Mimetype, Marshal: json.Marshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: xml.Mimetype, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}
}

// XMLMimetypes 返回以 XML 作为数据交换格式的配置项
func XMLMimetypes() []*web.Mimetype {
	return []*web.Mimetype{
		{Name: xml.Mimetype, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, Problem: xml.ProblemMimetype},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}
}

// JSONMimetypes 返回以 JSON 作为数据交换格式的配置项
func JSONMimetypes() []*web.Mimetype {
	return []*web.Mimetype{
		{Name: json.Mimetype, Marshal: json.Marshal, Unmarshal: json.Unmarshal, Problem: json.ProblemMimetype},
		{Name: "nil", Marshal: nil, Unmarshal: nil},
	}
}

// DefaultCompressions 提供当前框架内置的所有压缩算法
//
// contentType 指定所有算法应用的媒体类型，为空则表示对所有的内容都进行压缩。
func DefaultCompressions(contentType ...string) []*web.Compression {
	return []*web.Compression{
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
func BestSpeedCompressions(contentType ...string) []*web.Compression {
	return []*web.Compression{
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
func BestCompressionCompressions(contentType ...string) []*web.Compression {
	return []*web.Compression{
		{Compressor: compressor.NewGzipCompressor(gzip.BestCompression), Types: contentType},
		{Compressor: compressor.NewDeflateCompressor(flate.BestCompression, nil), Types: contentType},
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: contentType},
		{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestCompression}), Types: contentType},
		{Compressor: compressor.NewZstdCompressor(), Types: contentType},
	}
}
