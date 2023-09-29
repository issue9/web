// SPDX-License-Identifier: MIT

package compress

import (
	"bytes"
	"compress/lzw"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v3"
)

func BenchmarkGzip_Encoder(b *testing.B) {
	b.Run("gzip", func(b *testing.B) {
		benchCompressEncoder(b, NewGzipCompress(3))
	})

	b.Run("zstd", func(b *testing.B) {
		benchCompressEncoder(b, NewZstdCompress())
	})

	b.Run("deflate", func(b *testing.B) {
		benchCompressEncoder(b, NewDeflateCompress(3, nil))
	})

	b.Run("lzw", func(b *testing.B) {
		benchCompressEncoder(b, NewLZWCompress(lzw.LSB, 5))
	})

	b.Run("br", func(b *testing.B) {
		benchCompressEncoder(b, NewBrotliCompress(brotli.WriterOptions{}))
	})
}

func BenchmarkGzip_Decoder(b *testing.B) {
	a := assert.New(b, false)

	b.Run("gzip", func(b *testing.B) {
		c := NewGzipCompress(3)
		for i := 0; i < b.N; i++ {
			wc, err := c.Decoder(bytes.NewBuffer(gzipInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("zstd", func(b *testing.B) {
		c := NewZstdCompress()
		for i := 0; i < b.N; i++ {
			wc, err := c.Decoder(bytes.NewBuffer(zstdInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("deflate", func(b *testing.B) {
		c := NewDeflateCompress(3, nil)
		for i := 0; i < b.N; i++ {
			wc, err := c.Decoder(bytes.NewBuffer(deflateInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("lzw", func(b *testing.B) {
		c := NewLZWCompress(lzw.LSB, 5)
		for i := 0; i < b.N; i++ {
			wc, err := c.Decoder(bytes.NewBuffer(lzwInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("br", func(b *testing.B) {
		c := NewBrotliCompress(brotli.WriterOptions{})
		for i := 0; i < b.N; i++ {
			wc, err := c.Decoder(bytes.NewBuffer(brotliInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})
}

func benchCompressEncoder(b *testing.B, c Compress) {
	a := assert.New(b, false)
	w := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		w.Reset()

		wc, err := c.Encoder(w)
		a.NotError(err).
			NotNil(wc).
			NotError(wc.Close())
	}
}
