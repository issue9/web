// SPDX-License-Identifier: MIT

package app

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"

	"github.com/andybalholm/brotli"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/server"
)

var encodingFactory = map[string]enc{}

type enc struct {
	name string
	f    server.NewEncodingFunc
}

type encodingConfig struct {
	// Type content-type 的值
	//
	// 可以带通配符，比如 text/* 表示所有 text/ 开头的 content-type 都采用此压缩方法。
	Type string `json:"type" xml:"type,attr" yaml:"type"`

	// IDs 压缩方法的 ID 列表
	//
	// 这些 ID 值必须是由 RegisterEncoding 注册的，否则无效，默认情况下支持以下类型：
	// - deflate-default
	// - deflate-best-compression
	// - deflate-best-speed
	// - gzip-default
	// - gzip-best-compression
	// - gzip-best-speed
	// - compress-lsb-8
	// - compress-msb-8
	// - br-default
	// - br-best-compression
	// - br-best-speed
	IDs []string `json:"ids" xml:"id" yaml:"ids"`
}

func (conf *configOf[T]) sanitizeEncodings() *ConfigError {
	ids := make([]string, 0, len(encodingFactory))
	for _, e := range conf.Encodings {
		ids = append(ids, e.IDs...)
	}
	ids = sliceutil.Unique(ids, func(i, j string) bool { return i == j })

	conf.encodings = make(map[string]enc, len(ids))
	for _, id := range ids {
		item, found := encodingFactory[id]
		if !found {
			return &ConfigError{Message: localeutil.Error("%s not found", id), Field: "ids"}
		}
		conf.encodings[id] = item
	}
	return nil
}

func (conf *configOf[T]) buildEncodings(s *server.Server) *ConfigError {
	for id, item := range conf.encodings {
		s.AddEncoding(id, item.name, item.f)
	}

	for _, enc := range conf.Encodings {
		s.AllowEncoding(enc.Type, enc.IDs...)
	}

	return nil
}

// RegisterEncoding 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
// name 表示此压缩方法的名称，可以相同；
// f 生成压缩对象的方法；
func RegisterEncoding(id, name string, f server.NewEncodingFunc) {
	if _, found := encodingFactory[id]; found {
		panic("已经存在相同的 id:" + id)
	}
	encodingFactory[id] = enc{name: name, f: f}
}

func init() {
	RegisterEncoding("deflate-default", "deflate", encoding.DeflateWriter(flate.DefaultCompression))
	RegisterEncoding("deflate-best-compression", "deflate", encoding.DeflateWriter(flate.BestCompression))
	RegisterEncoding("deflate-best-speed", "deflate", encoding.DeflateWriter(flate.BestSpeed))

	RegisterEncoding("gzip-default", "gzip", encoding.GZipWriter(gzip.DefaultCompression))
	RegisterEncoding("gzip-best-compression", "gzip", encoding.GZipWriter(gzip.BestCompression))
	RegisterEncoding("gzip-best-speed", "gzip", encoding.GZipWriter(gzip.BestSpeed))

	RegisterEncoding("compress-lsb-8", "compress", encoding.CompressWriter(lzw.LSB, 8))
	RegisterEncoding("compress-msb-8", "compress", encoding.CompressWriter(lzw.MSB, 8))

	RegisterEncoding("br-default", "br", encoding.BrotliWriter(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterEncoding("br-best-compression", "br", encoding.BrotliWriter(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterEncoding("br-best-speed", "br", encoding.BrotliWriter(brotli.WriterOptions{Quality: brotli.BestSpeed}))
}
