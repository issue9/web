// SPDX-License-Identifier: MIT

package codec

import (
	"bytes"
	"compress/lzw"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v3"

	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/codec/mimetype/xml"
)

func BenchmarkCodec_Accept(b *testing.B) {
	a := assert.New(b, false)
	mt := New()
	a.NotNil(mt)

	mt.AddMimetype("font/wottf", xml.BuildMarshal, xml.Unmarshal, "")
	mt.AddMimetype("text/plain", json.BuildMarshal, json.Unmarshal, "text/plain+problem")

	for i := 0; i < b.N; i++ {
		item := mt.Accept("font/wottf;q=0.9")
		a.NotNil(item)
	}
}

func BenchmarkCodec_ContentType(b *testing.B) {
	a := assert.New(b, false)
	mt := New()
	a.NotNil(mt)

	mt.AddMimetype("font/1", xml.BuildMarshal, xml.Unmarshal, "")
	mt.AddMimetype("font/2", xml.BuildMarshal, xml.Unmarshal, "")
	mt.AddMimetype("font/3", xml.BuildMarshal, xml.Unmarshal, "")

	b.Run("charset=utf-8", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			marshal, encoding, err := mt.ContentType("font/2;charset=utf-8")
			a.NotError(err).NotNil(marshal).Nil(encoding)
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			marshal, encoding, err := mt.ContentType("font/2;charset=gbk")
			a.NotError(err).NotNil(marshal).NotNil(encoding)
		}
	})
}

func BenchmarkCodec_ContentEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := New().AddCompressor("gzip", NewGzipCompressor(3), "application/*")
		for i := 0; i < b.N; i++ {
			r := bytes.NewBuffer(gzipInitData)
			_, err := c.ContentEncoding("gzip", r)
			a.NotError(err)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := New().
			AddCompressor("gzip", NewGzipCompressor(3), "application/*").
			AddCompressor("br", NewBrotliCompressor(brotli.WriterOptions{}), "text/*").
			AddCompressor("deflate", NewDeflateCompressor(3, nil), "image/*").
			AddCompressor("zstd", NewZstdCompressor(), "text/html").
			AddCompressor("compress", NewLZWCompressor(lzw.LSB, 8), "text/plain")

		for i := 0; i < b.N; i++ {
			r := bytes.NewBuffer(gzipInitData)
			_, err := c.ContentEncoding("gzip", r)
			a.NotError(err)
		}
	})
}

func BenchmarkCodec_AcceptEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := New().AddCompressor("gzip", NewGzipCompressor(3), "application/*")
		for i := 0; i < b.N; i++ {
			_, _, na := c.AcceptEncoding("application/json", "gzip", nil)
			a.False(na)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := New().
			AddCompressor("gzip", NewGzipCompressor(3), "application/*").
			AddCompressor("br", NewBrotliCompressor(brotli.WriterOptions{}), "text/*").
			AddCompressor("deflate", NewDeflateCompressor(3, nil), "image/*").
			AddCompressor("zstd", NewZstdCompressor(), "text/html").
			AddCompressor("compress", NewLZWCompressor(lzw.LSB, 8), "text/plain")

		for i := 0; i < b.N; i++ {
			_, _, na := c.AcceptEncoding("text/plain", "compress", nil)
			a.False(na)
		}
	})
}

func BenchmarkCodec_getMatchCompresses(b *testing.B) {
	c := New().
		AddCompressor("gzip", NewGzipCompressor(3), "application/*").
		AddCompressor("br", NewBrotliCompressor(brotli.WriterOptions{}), "text/*").
		AddCompressor("deflate", NewDeflateCompressor(3, nil), "image/*").
		AddCompressor("zstd", NewZstdCompressor(), "text/html").
		AddCompressor("compress", NewLZWCompressor(lzw.LSB, 8), "text/plain")

	for i := 0; i < b.N; i++ {
		c.getMatchCompresses("text/plan")
	}
}

func BenchmarkCodec_NewEncoder(b *testing.B) {
	b.Run("gzip", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewGzipCompressor(3))
	})

	b.Run("zstd", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewZstdCompressor())
	})

	b.Run("deflate", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewDeflateCompressor(3, nil))
	})

	b.Run("lzw", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewLZWCompressor(lzw.LSB, 5))
	})

	b.Run("br", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewBrotliCompressor(brotli.WriterOptions{}))
	})
}

func BenchmarkCodec_NewDecoder(b *testing.B) {
	a := assert.New(b, false)

	b.Run("gzip", func(b *testing.B) {
		c := NewGzipCompressor(3)
		for i := 0; i < b.N; i++ {
			wc, err := c.NewDecoder(bytes.NewBuffer(gzipInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("zstd", func(b *testing.B) {
		c := NewZstdCompressor()
		for i := 0; i < b.N; i++ {
			wc, err := c.NewDecoder(bytes.NewBuffer(zstdInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("deflate", func(b *testing.B) {
		c := NewDeflateCompressor(3, nil)
		for i := 0; i < b.N; i++ {
			wc, err := c.NewDecoder(bytes.NewBuffer(deflateInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("lzw", func(b *testing.B) {
		c := NewLZWCompressor(lzw.LSB, 5)
		for i := 0; i < b.N; i++ {
			wc, err := c.NewDecoder(bytes.NewBuffer(lzwInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("br", func(b *testing.B) {
		c := NewBrotliCompressor(brotli.WriterOptions{})
		for i := 0; i < b.N; i++ {
			wc, err := c.NewDecoder(bytes.NewBuffer(brotliInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})
}

func benchCompressor_NewEncoder(b *testing.B, c Compressor) {
	a := assert.New(b, false)
	w := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		w.Reset()

		wc, err := c.NewEncoder(w)
		a.NotError(err).
			NotNil(wc).
			NotError(wc.Close())
	}
}
