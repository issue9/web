// SPDX-License-Identifier: MIT

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
)

type (
	// Compress 压缩算法的接口
	Compress interface {
		// Decoder 将 r 包装成为当前压缩算法的解码器
		Decoder(r io.Reader) (io.ReadCloser, error)

		// Encoder 将 w 包装成当前压缩算法的编码器
		Encoder(w io.Writer) (io.WriteCloser, error)
	}

	zstdCompress struct {
		readers *sync.Pool
		writers *sync.Pool
	}

	brotliCompress struct {
		readers *sync.Pool
		writers *sync.Pool
	}

	lzwCompress struct {
		order   lzw.Order
		width   int
		readers *sync.Pool
		writers *sync.Pool
	}

	gzipCompress struct {
		readInitData []byte
		readers      *sync.Pool
		writers      *sync.Pool
	}

	deflateCompress struct {
		dict    []byte
		readers *sync.Pool
		writers *sync.Pool
	}

	// Named 带名称的压缩算法
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
// [zstd]: https://www.rfc-editor.org/rfc/rfc8878.html
func NewZstdCompress(do []zstd.DOption, eo []zstd.EOption) Compress {
	return &zstdCompress{
		readers: &sync.Pool{New: func() any {
			r, err := zstd.NewReader(nil, do...)
			if err != nil {
				panic(err)
			}
			return r
		}},

		writers: &sync.Pool{New: func() any {
			w, err := zstd.NewWriter(nil, eo...)
			if err != nil {
				panic(err)
			}
			return w
		}},
	}
}

func (c *zstdCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := c.readers.Get().(*zstd.Decoder)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(io.NopCloser(rr), func() { c.readers.Put(rr) }), nil
}

func (c *zstdCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*zstd.Encoder)
	ww.Reset(w)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

// NewBrotliCompress 声明基于 [br] 的压缩算法
//
// [br]: https://www.rfc-editor.org/rfc/rfc7932.html
func NewBrotliCompress(o brotli.WriterOptions) Compress {
	return &brotliCompress{
		readers: &sync.Pool{New: func() any {
			return brotli.NewReader(nil)
		}},

		writers: &sync.Pool{New: func() any {
			return brotli.NewWriterOptions(nil, o)
		}},
	}
}

func (c *brotliCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := c.readers.Get().(*brotli.Reader)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(io.NopCloser(rr), func() { c.readers.Put(rr) }), nil
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
	return &lzwCompress{
		order: order,
		width: width,

		readers: &sync.Pool{New: func() any {
			return lzw.NewReader(nil, order, width)
		}},

		writers: &sync.Pool{New: func() any {
			return lzw.NewWriter(nil, order, width)
		}},
	}
}

func (c *lzwCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := c.readers.Get().(*lzw.Reader)
	rr.Reset(r, c.order, c.width)
	return wrapDecoder(rr, func() { c.readers.Put(rr) }), nil
}

func (c *lzwCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*lzw.Writer)
	ww.Reset(w, c.order, c.width)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

// NewGzipCompress 声明基于 gzip 的压缩算法
func NewGzipCompress(level int) Compress {
	r := &bytes.Buffer{}
	gw := gzip.NewWriter(r)
	gw.Write([]byte(""))
	gw.Flush()
	gw.Close()
	data := r.Bytes()

	return &gzipCompress{
		readInitData: data,

		readers: &sync.Pool{New: func() any {
			r, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				panic(err)
			}
			return r
		}},
		writers: &sync.Pool{New: func() any {
			w, err := gzip.NewWriterLevel(nil, level)
			if err != nil {
				panic(err)
			}
			return w
		}},
	}
}

func (c *gzipCompress) Decoder(r io.Reader) (io.ReadCloser, error) {
	rr := c.readers.Get().(*gzip.Reader)
	if err := rr.Reset(r); err != nil {
		return nil, err
	}
	return wrapDecoder(rr, func() { c.readers.Put(rr) }), nil
}

func (c *gzipCompress) Encoder(w io.Writer) (io.WriteCloser, error) {
	ww := c.writers.Get().(*gzip.Writer)
	ww.Reset(w)
	return wrapEncoder(ww, func() { c.writers.Put(ww) }), nil
}

// NewDeflateCompress 声明基于 deflate 的压缩算法
func NewDeflateCompress(level int, dict []byte) Compress {
	return &deflateCompress{
		dict: dict,

		readers: &sync.Pool{New: func() any {
			return flate.NewReaderDict(nil, dict)
		}},

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
	rr := c.readers.Get().(flate.Resetter)
	if err := rr.Reset(r, c.dict); err != nil {
		return nil, err
	}
	return wrapDecoder(rr.(io.ReadCloser), func() { c.readers.Put(rr) }), nil
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
	return &encoder{WriteCloser: w, destroy: f}
}

func wrapDecoder(r io.ReadCloser, f func()) *decoder {
	return &decoder{ReadCloser: r, destroy: f}
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
