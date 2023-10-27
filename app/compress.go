// SPDX-License-Identifier: MIT

package app

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"

	"github.com/issue9/web"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/codec/compressor"
)

var compressorFactory = map[string]enc{}

type enc struct {
	name string
	c    compressor.Compressor
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

func (conf *configOf[T]) sanitizeCompresses() *web.FieldError {
	conf.compressors = make([]*codec.Compression, 0, len(conf.Compressors))
	for index, e := range conf.Compressors {
		enc, found := compressorFactory[e.ID]
		if !found {
			field := "compresses[" + strconv.Itoa(index) + "].id"
			return web.NewFieldError(field, web.NewLocaleError("%s not found", e.ID))
		}

		conf.compressors = append(conf.compressors, &codec.Compression{
			Compressor: enc.c,
			Types:      e.Types,
		})
	}
	return nil
}

// RegisterCompression 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
// name 表示此压缩方法的名称，可以相同；
// f 生成压缩对象的方法；
func RegisterCompression(id, name string, f compressor.Compressor) {
	if _, found := compressorFactory[id]; found {
		panic("已经存在相同的 id:" + id)
	}
	compressorFactory[id] = enc{name: name, c: f}
}

func init() {
	RegisterCompression("deflate-default", "deflate", compressor.NewDeflateCompressor(flate.DefaultCompression, nil))
	RegisterCompression("deflate-best-compression", "deflate", compressor.NewDeflateCompressor(flate.BestCompression, nil))
	RegisterCompression("deflate-best-speed", "deflate", compressor.NewDeflateCompressor(flate.BestSpeed, nil))

	RegisterCompression("gzip-default", "gzip", compressor.NewGzipCompressor(gzip.DefaultCompression))
	RegisterCompression("gzip-best-compression", "gzip", compressor.NewGzipCompressor(gzip.BestCompression))
	RegisterCompression("gzip-best-speed", "gzip", compressor.NewGzipCompressor(gzip.BestSpeed))

	RegisterCompression("compress-lsb-8", "compress", compressor.NewLZWCompressor(lzw.LSB, 8))
	RegisterCompression("compress-msb-8", "compress", compressor.NewLZWCompressor(lzw.MSB, 8))

	RegisterCompression("br-default", "br", compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterCompression("br-best-compression", "br", compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterCompression("br-best-speed", "br", compressor.NewBrotliCompressor(brotli.WriterOptions{Quality: brotli.BestSpeed}))

	RegisterCompression("zstd-default", "zstd", compressor.NewZstdCompressor())
}
