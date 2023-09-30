// SPDX-License-Identifier: MIT

// Package compress 提供压缩算法相关的功能
package compress

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
	// Compress 压缩算法的接口
	Compress interface {
		// Decoder 将 r 包装成为当前压缩算法的解码器
		Decoder(r io.Reader) (io.ReadCloser, error)

		// Encoder 将 w 包装成当前压缩算法的编码器
		Encoder(w io.Writer) (io.WriteCloser, error)
	}

	zstdCompress struct{}

	brotliCompress struct {
		writers *sync.Pool
	}

	lzwCompress struct {
		order lzw.Order
		width int
	}

	gzipCompress struct{}

	deflateCompress struct {
		dict    []byte
		writers *sync.Pool
	}

	// NamedCompress 带名称的压缩算法
	NamedCompress struct {
		name     string
		compress Compress

		// contentType 是具体值的，比如 text/xml
		allowTypes []string

		// contentType 是模糊类型的，比如 text/*，
		// 只有在 allowTypes 找不到时，才在此处查找。
		allowTypesPrefix []string
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

// NewZstdCompress 声明基于 [zstd] 的压缩算法
//
// NOTE: 如果是针对浏览器的，请注意[浏览器支持情况]
//
// [浏览器支持情况]: https://caniuse.com/zstd
// [zstd]: https://www.rfc-editor.org/rfc/rfc8878.html
func NewZstdCompress() Compress {
	return &zstdCompress{} // TODO: 替换为官方的 https://github.com/golang/go/issues/62513
}

func (c *zstdCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := zstdReaders.Get().(*zstd.Decoder)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(io.NopCloser(rr), func() { zstdReaders.Put(rr) }), nil
}

func (c *zstdCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := zstdWriters.Get().(*zstd.Encoder)
	ww.Reset(w)
	return wrapEncoder(ww, func() { zstdWriters.Put(ww) }), nil
}

// NewBrotliCompress 声明基于 [br] 的压缩算法
//
// [br]: https://www.rfc-editor.org/rfc/rfc7932.html
func NewBrotliCompress(o brotli.WriterOptions) Compress {
	return &brotliCompress{
		writers: &sync.Pool{New: func() any {
			return brotli.NewWriterOptions(nil, o)
		}},
	}
}

func (c *brotliCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := brotliReaders.Get().(*brotli.Reader)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(io.NopCloser(rr), func() { brotliReaders.Put(rr) }), nil
}

func (c *brotliCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*brotli.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

// NewLZWCompress 声明基于 lzw 的压缩算法
//
// NOTE: 在 http 报头中名称为 compress 或是 x-compress
func NewLZWCompress(order lzw.Order, width int) Compress {
	return &lzwCompress{order: order, width: width}
}

func (c *lzwCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := lzwReaders.Get().(*lzw.Reader)
	rr.Reset(r, c.order, c.width)
	return wrapDecoder(rr, func() { lzwReaders.Put(rr) }), nil
}

func (c *lzwCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := lzwWriters.Get().(*lzw.Writer)
	ww.Reset(w, c.order, c.width)
	return wrapEncoder(ww, func() { lzwWriters.Put(ww) }), nil
}

// NewGzipCompress 声明基于 gzip 的压缩算法
func NewGzipCompress(level int) Compress { return &gzipCompress{} }

func (c *gzipCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := gzipReaders.Get().(*gzip.Reader)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(rr, func() { gzipReaders.Put(rr) }), nil
}

func (c *gzipCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := gzipWriters.Get().(*gzip.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { gzipWriters.Put(ww) }), nil
}

// NewDeflateCompress 声明基于 deflate 的压缩算法
func NewDeflateCompress(level int, dict []byte) Compress {
	return &deflateCompress{
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

func (c *deflateCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := deflateReaders.Get().(flate.Resetter)
	if err := rr.Reset(r, c.dict); err != nil {
		return nil, err
	}
	return wrapDecoder(rr.(io.ReadCloser), func() { deflateReaders.Put(rr) }), nil
}

func (c *deflateCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*flate.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

// Name 算法名称
func (c *NamedCompress) Name() string { return c.name }

// Compress 关联的压缩算法接口
func (c *NamedCompress) Compress() Compress { return c.compress }

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
