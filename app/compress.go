// SPDX-License-Identifier: MIT

package app

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"

	"github.com/issue9/web"
)

var compressorFactory = map[string]enc{}

type enc struct {
	name string
	c    web.Compressor
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
	conf.compresses = make([]*web.Compress, 0, len(conf.Compresses))
	for index, e := range conf.Compresses {
		enc, found := compressorFactory[e.ID]
		if !found {
			field := "compresses[" + strconv.Itoa(index) + "].id"
			return web.NewFieldError(field, web.NewLocaleError("%s not found", e.ID))
		}

		conf.compresses = append(conf.compresses, &web.Compress{
			Name:       enc.name,
			Compressor: enc.c,
			Types:      e.Types,
		})
	}
	return nil
}

// RegisterCompress 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
// name 表示此压缩方法的名称，可以相同；
// f 生成压缩对象的方法；
func RegisterCompress(id, name string, f web.Compressor) {
	if _, found := compressorFactory[id]; found {
		panic("已经存在相同的 id:" + id)
	}
	compressorFactory[id] = enc{name: name, c: f}
}

func init() {
	RegisterCompress("deflate-default", "deflate", web.NewDeflateCompress(flate.DefaultCompression, nil))
	RegisterCompress("deflate-best-compression", "deflate", web.NewDeflateCompress(flate.BestCompression, nil))
	RegisterCompress("deflate-best-speed", "deflate", web.NewDeflateCompress(flate.BestSpeed, nil))

	RegisterCompress("gzip-default", "gzip", web.NewGzipCompress(gzip.DefaultCompression))
	RegisterCompress("gzip-best-compression", "gzip", web.NewGzipCompress(gzip.BestCompression))
	RegisterCompress("gzip-best-speed", "gzip", web.NewGzipCompress(gzip.BestSpeed))

	RegisterCompress("compress-lsb-8", "compress", web.NewLZWCompress(lzw.LSB, 8))
	RegisterCompress("compress-msb-8", "compress", web.NewLZWCompress(lzw.MSB, 8))

	RegisterCompress("br-default", "br", web.NewBrotliCompress(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterCompress("br-best-compression", "br", web.NewBrotliCompress(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterCompress("br-best-speed", "br", web.NewBrotliCompress(brotli.WriterOptions{Quality: brotli.BestSpeed}))

	RegisterCompress("zstd-default", "zstd", web.NewZstdCompress())
}
