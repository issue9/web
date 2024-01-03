// SPDX-License-Identifier: MIT

//go:generate go run ./make_data.go

// Package compressor 提供了压缩算法的实现
package compressor

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

var (
	decoderPool = &sync.Pool{New: func() any { return &decoder{} }}
	encoderPool = &sync.Pool{New: func() any { return &encoder{} }}

	// zstd
	zstdReaders = &sync.Pool{New: func() any {
		r, err := zstd.NewReader(bytes.NewReader(zstdInitData))
		if err != nil {
			panic(err)
		}
		return r
	}}
	zstdWriters = &sync.Pool{New: func() any {
		w, err := zstd.NewWriter(nil)
		if err != nil {
			panic(err)
		}
		return w
	}}

	// brotli
	brotliReaders = &sync.Pool{New: func() any {
		return brotli.NewReader(bytes.NewReader(brotliInitData))
	}}

	// lzw
	lzwReaders = &sync.Pool{New: func() any {
		return lzw.NewReader(bytes.NewReader(lzwInitData), lzw.LSB, 8)
	}}
	lzwWriters = &sync.Pool{New: func() any {
		return lzw.NewWriter(nil, lzw.LSB, 8)
	}}

	// gzip
	gzipReaders = &sync.Pool{New: func() any {
		r, err := gzip.NewReader(bytes.NewReader(gzipInitData))
		if err != nil {
			panic(err)
		}
		return r
	}}
	gzipWriters = &sync.Pool{New: func() any { return gzip.NewWriter(nil) }}

	// deflate
	deflateReaders = &sync.Pool{New: func() any {
		return flate.NewReader(bytes.NewReader(deflateInitData))
	}}
)

type (
	// Compressor 压缩算法的接口
	Compressor interface {
		// Name 算法的名称
		Name() string

		// NewDecoder 将 r 包装成为当前压缩算法的解码器
		NewDecoder(r io.Reader) (io.ReadCloser, error)

		// NewEncoder 将 w 包装成当前压缩算法的编码器
		NewEncoder(w io.Writer) (io.WriteCloser, error)
	}

	zstdCompressor struct{}

	brotliCompressor struct {
		writers *sync.Pool
	}

	lzwCompressor struct {
		order lzw.Order
		width int
	}

	gzipCompressor struct{}

	deflateCompressor struct {
		dict    []byte
		writers *sync.Pool
	}

	decoder struct {
		io.ReadCloser
		destroy func()
	}

	encoder struct {
		io.WriteCloser
		destroy func()
	}
)

// NewZstd 声明基于 [zstd] 的压缩算法
//
// NOTE: 请注意[浏览器支持情况]
//
// [浏览器支持情况]: https://caniuse.com/zstd
// [zstd]: https://www.rfc-editor.org/rfc/rfc8878.html
func NewZstd() Compressor {
	return &zstdCompressor{} // TODO: 替换为官方的 https://github.com/golang/go/issues/62513
}

func (c *zstdCompressor) Name() string { return "zstd" }

func (c *zstdCompressor) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	rr := zstdReaders.Get().(*zstd.Decoder)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(io.NopCloser(rr), func() { zstdReaders.Put(rr) }), nil
}

func (c *zstdCompressor) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	ww := zstdWriters.Get().(*zstd.Encoder)
	ww.Reset(w)
	return wrapEncoder(ww, func() { zstdWriters.Put(ww) }), nil
}

// NewBrotli 声明基于 [br] 的压缩算法
//
// [br]: https://www.rfc-editor.org/rfc/rfc7932.html
func NewBrotli(o brotli.WriterOptions) Compressor {
	return &brotliCompressor{
		writers: &sync.Pool{New: func() any {
			return brotli.NewWriterOptions(nil, o)
		}},
	}
}

func (c *brotliCompressor) Name() string { return "br" }

func (c *brotliCompressor) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	rr := brotliReaders.Get().(*brotli.Reader)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(io.NopCloser(rr), func() { brotliReaders.Put(rr) }), nil
}

func (c *brotliCompressor) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*brotli.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

// NewLZW 声明基于 lzw 的压缩算法
//
// NOTE: 在 http 报头中名称为 compress 或是 x-compress
func NewLZW(order lzw.Order, width int) Compressor {
	return &lzwCompressor{order: order, width: width}
}

func (c *lzwCompressor) Name() string { return "compress" }

func (c *lzwCompressor) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	rr := lzwReaders.Get().(*lzw.Reader)
	rr.Reset(r, c.order, c.width)
	return wrapDecoder(rr, func() { lzwReaders.Put(rr) }), nil
}

func (c *lzwCompressor) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	ww := lzwWriters.Get().(*lzw.Writer)
	ww.Reset(w, c.order, c.width)
	return wrapEncoder(ww, func() { lzwWriters.Put(ww) }), nil
}

// NewGzip 声明基于 gzip 的压缩算法
func NewGzip(level int) Compressor { return &gzipCompressor{} }

func (c *gzipCompressor) Name() string { return "gzip" }

func (c *gzipCompressor) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	rr := gzipReaders.Get().(*gzip.Reader)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(rr, func() { gzipReaders.Put(rr) }), nil
}

func (c *gzipCompressor) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	ww := gzipWriters.Get().(*gzip.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { gzipWriters.Put(ww) }), nil
}

// NewDeflate 声明基于 deflate 的压缩算法
func NewDeflate(level int, dict []byte) Compressor {
	return &deflateCompressor{
		dict: dict,

		writers: &sync.Pool{New: func() any {
			var w *flate.Writer
			var err error
			if len(dict) == 0 {
				w, err = flate.NewWriter(nil, level)
			} else {
				w, err = flate.NewWriterDict(nil, level, dict)
			}

			if err != nil { // NewWriter 就一个判断参数错误的，更应该像是 panic
				panic(err)
			}
			return w
		}},
	}
}

func (c *deflateCompressor) Name() string { return "deflate" }

func (c *deflateCompressor) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	rr := deflateReaders.Get().(flate.Resetter)
	if err := rr.Reset(r, c.dict); err != nil {
		return nil, err
	}
	return wrapDecoder(rr.(io.ReadCloser), func() { deflateReaders.Put(rr) }), nil
}

func (c *deflateCompressor) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*flate.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

func wrapEncoder(w io.WriteCloser, f func()) *encoder {
	e := encoderPool.Get().(*encoder)
	e.WriteCloser = w
	e.destroy = f
	return e
}

func wrapDecoder(r io.ReadCloser, f func()) *decoder {
	d := decoderPool.Get().(*decoder)
	d.ReadCloser = r
	d.destroy = f
	return d
}

func (e *encoder) Close() error {
	e.destroy()
	err := e.WriteCloser.Close()
	encoderPool.Put(e)
	return err
}

func (d *decoder) Close() error {
	d.destroy()
	err := d.ReadCloser.Close()
	decoderPool.Put(d)
	return err
}
