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
	mt, err := New("ms", "cs", APIMimetypes(), nil)
	a.NotError(err).NotNil(mt)

	for i := 0; i < b.N; i++ {
		item := mt.Accept("application/json;q=0.9")
		a.NotNil(item)
	}
}

func BenchmarkCodec_ContentType(b *testing.B) {
	a := assert.New(b, false)
	mt, err := New("ms", "cs", APIMimetypes(), nil)
	a.NotError(err).NotNil(mt)

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

		c, err := New("ms", "cs", nil, []*Compression{
			{Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
		})
		a.NotError(err).NotNil(c)

		for i := 0; i < b.N; i++ {
			r := bytes.NewBuffer([]byte{})
			_, err := c.ContentEncoding("zstd", r)
			a.NotError(err)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c, err := New("ms", "cs", nil, []*Compression{
			{Compressor: compressor.NewGzipCompressor(3), Types: []string{"application/*"}},
			{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: []string{"text/*"}},
			{Compressor: compressor.NewDeflateCompressor(3, nil), Types: []string{"image/*"}},
			{Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
			{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: []string{"text/plain"}},
		})
		a.NotError(err).NotNil(c)

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

		c, err := New("ms", "cs", nil, []*Compression{
			{Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
		})
		a.NotError(err).NotNil(c)

		for i := 0; i < b.N; i++ {
			_, _, na := c.AcceptEncoding("application/json", "zstd", nil)
			a.False(na)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c, err := New("ms", "cs", nil, []*Compression{
			{Compressor: compressor.NewGzipCompressor(3), Types: []string{"application/*"}},
			{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: []string{"text/*"}},
			{Compressor: compressor.NewDeflateCompressor(3, nil), Types: []string{"image/*"}},
			{Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
			{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: []string{"text/plain"}},
		})
		a.NotError(err).NotNil(c)

		for i := 0; i < b.N; i++ {
			_, _, na := c.AcceptEncoding("text/plain", "compress", nil)
			a.False(na)
		}
	})
}

func BenchmarkCodec_getMatchCompresses(b *testing.B) {
	a := assert.New(b, false)

	cc, err := New("ms", "cs", nil, []*Compression{
		{Compressor: compressor.NewGzipCompressor(3), Types: []string{"application/*"}},
		{Compressor: compressor.NewBrotliCompressor(brotli.WriterOptions{}), Types: []string{"text/*"}},
		{Compressor: compressor.NewDeflateCompressor(3, nil), Types: []string{"image/*"}},
		{Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 8), Types: []string{"text/plain"}},
	})
	a.NotError(err).NotNil(cc)
	c := cc.(*codec)

	for i := 0; i < b.N; i++ {
		c.getMatchCompresses("text/plan")
	}
}
