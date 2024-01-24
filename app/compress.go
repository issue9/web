// SPDX-License-Identifier: MIT

package app

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"

	"github.com/issue9/web"
	"github.com/issue9/web/compressor"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

var compressorFactory = newRegister[compressor.Compressor]()

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
	conf.compressors = make([]*server.Compression, 0, len(conf.Compressors))
	for index, e := range conf.Compressors {
		enc, found := compressorFactory.get(e.ID)
		if !found {
			field := "compresses[" + strconv.Itoa(index) + "].id"
			return web.NewFieldError(field, locales.ErrNotFound(e.ID))
		}

		conf.compressors = append(conf.compressors, &server.Compression{
			Compressor: enc,
			Types:      e.Types,
		})
	}
	return nil
}

// RegisterCompression 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
// c 压缩算法；
func RegisterCompression(id string, c compressor.Compressor) { compressorFactory.register(c, id) }

func init() {
	RegisterCompression("deflate-default", compressor.NewDeflate(flate.DefaultCompression, nil))
	RegisterCompression("deflate-best-compression", compressor.NewDeflate(flate.BestCompression, nil))
	RegisterCompression("deflate-best-speed", compressor.NewDeflate(flate.BestSpeed, nil))

	RegisterCompression("gzip-default", compressor.NewGzip(gzip.DefaultCompression))
	RegisterCompression("gzip-best-compression", compressor.NewGzip(gzip.BestCompression))
	RegisterCompression("gzip-best-speed", compressor.NewGzip(gzip.BestSpeed))

	RegisterCompression("compress-lsb-8", compressor.NewLZW(lzw.LSB, 8))
	RegisterCompression("compress-msb-8", compressor.NewLZW(lzw.MSB, 8))

	RegisterCompression("br-default", compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterCompression("br-best-compression", compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterCompression("br-best-speed", compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.BestSpeed}))

	RegisterCompression("zstd-default", compressor.NewZstd())
}
