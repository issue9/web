// SPDX-License-Identifier: MIT

package app

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"

	"github.com/issue9/web"
	"github.com/issue9/web/compress"
)

var encodingFactory = map[string]enc{}

type enc struct {
	name string
	f    compress.Compress
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
		enc, found := encodingFactory[e.ID]
		if !found {
			field := "compresses[" + strconv.Itoa(index) + "].id"
			return web.NewFieldError(field, web.NewLocaleError("%s not found", e.ID))
		}

		conf.compresses = append(conf.compresses, &web.Compress{
			Name:     enc.name,
			Compress: enc.f,
			Types:    e.Types,
		})
	}
	return nil
}

// RegisterCompress 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
// name 表示此压缩方法的名称，可以相同；
// f 生成压缩对象的方法；
func RegisterCompress(id, name string, f compress.Compress) {
	if _, found := encodingFactory[id]; found {
		panic("已经存在相同的 id:" + id)
	}
	encodingFactory[id] = enc{name: name, f: f}
}

func init() {
	RegisterCompress("deflate-default", "deflate", compress.NewDeflateCompress(flate.DefaultCompression, nil))
	RegisterCompress("deflate-best-compression", "deflate", compress.NewDeflateCompress(flate.BestCompression, nil))
	RegisterCompress("deflate-best-speed", "deflate", compress.NewDeflateCompress(flate.BestSpeed, nil))

	RegisterCompress("gzip-default", "gzip", compress.NewGzipCompress(gzip.DefaultCompression))
	RegisterCompress("gzip-best-compression", "gzip", compress.NewGzipCompress(gzip.BestCompression))
	RegisterCompress("gzip-best-speed", "gzip", compress.NewGzipCompress(gzip.BestSpeed))

	RegisterCompress("compress-lsb-8", "compress", compress.NewLZWCompress(lzw.LSB, 8))
	RegisterCompress("compress-msb-8", "compress", compress.NewLZWCompress(lzw.MSB, 8))

	RegisterCompress("br-default", "br", compress.NewBrotliCompress(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterCompress("br-best-compression", "br", compress.NewBrotliCompress(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterCompress("br-best-speed", "br", compress.NewBrotliCompress(brotli.WriterOptions{Quality: brotli.BestSpeed}))

	RegisterCompress("zstd-default", "zstd", compress.NewZstdCompress())
}
