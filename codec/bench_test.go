// SPDX-License-Identifier: MIT

package codec

import (
	"bytes"
	"compress/lzw"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v3"
	"github.com/issue9/web/codec/compressor"
)

func BenchmarkCodec_Accept(b *testing.B) {
	a := assert.New(b, false)
	mt := New(APIMimetypes(), nil)
	a.NotNil(mt)

	for i := 0; i < b.N; i++ {
		item := mt.Accept("application/json;q=0.9")
		a.NotNil(item)
	}
}

func BenchmarkCodec_ContentType(b *testing.B) {
	a := assert.New(b, false)
	mt := New(APIMimetypes(), nil)
	a.NotNil(mt)

	b.Run("charset=utf-8", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			marshal, encoding, err := mt.ContentType("application/xml;charset=utf-8")
			a.NotError(err).NotNil(marshal).Nil(encoding)
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			marshal, encoding, err := mt.ContentType("application/xml;charset=gbk")
			a.NotError(err).NotNil(marshal).NotNil(encoding)
		}
	})
}

func BenchmarkCodec_ContentEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := New(nil, []*Compression{
			{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
		})
		for i := 0; i < b.N; i++ {
			r := bytes.NewBuffer([]byte{})
			_, err := c.ContentEncoding("zstd", r)
			a.NotError(err)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := New(nil, []*Compression{
			{Name: "gzip", Compressor: compressor.NewGzipCompressor(3), Types: []string{"application/*"}},
			{Name: "br", Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: []string{"text/*"}},
			{Name: "deflate", Compressor: compressor.NewDeflateCompressor(3, nil), Types: []string{"image/*"}},
			{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
			{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: []string{"text/plain"}},
		})
		for i := 0; i < b.N; i++ {
			r := bytes.NewBuffer([]byte{})
			_, err := c.ContentEncoding("zstd", r)
			a.NotError(err)
		}
	})
}

func BenchmarkCodec_AcceptEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := New(nil, []*Compression{
			{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
		})
		for i := 0; i < b.N; i++ {
			_, _, na := c.AcceptEncoding("application/json", "zstd", nil)
			a.False(na)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := New(nil, []*Compression{
			{Name: "gzip", Compressor: compressor.NewGzipCompressor(3), Types: []string{"application/*"}},
			{Name: "br", Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: []string{"text/*"}},
			{Name: "deflate", Compressor: compressor.NewDeflateCompressor(3, nil), Types: []string{"image/*"}},
			{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
			{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: []string{"text/plain"}},
		})

		for i := 0; i < b.N; i++ {
			_, _, na := c.AcceptEncoding("text/plain", "compress", nil)
			a.False(na)
		}
	})
}

func BenchmarkCodec_getMatchCompresses(b *testing.B) {
	cc := New(nil, []*Compression{
		{Name: "gzip", Compressor: compressor.NewGzipCompressor(3), Types: []string{"application/*"}},
		{Name: "br", Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: []string{"text/*"}},
		{Name: "deflate", Compressor: compressor.NewDeflateCompressor(3, nil), Types: []string{"image/*"}},
		{Name: "zstd", Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
		{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: []string{"text/plain"}},
	})
	c := cc.(*codec)

	for i := 0; i < b.N; i++ {
		c.getMatchCompresses("text/plan")
	}
}
