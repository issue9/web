// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package codec

import (
	"bytes"
	"compress/lzw"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v6"
)

func BenchmarkCodec_NewEncoder(b *testing.B) {
	b.ReportAllocs()

	b.Run("gzip", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewGzip(3))
	})

	b.Run("zstd", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewZstd())
	})

	b.Run("deflate", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewDeflate(3, nil))
	})

	b.Run("lzw", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewLZW(lzw.LSB, 5))
	})

	b.Run("br", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewBrotli(brotli.WriterOptions{}))
	})
}

func BenchmarkCodec_NewDecoder(b *testing.B) {
	a := assert.New(b, false)

	b.Run("gzip", func(b *testing.B) {
		c := NewGzip(3)
		b.ReportAllocs()
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(gzipInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("zstd", func(b *testing.B) {
		c := NewZstd()
		b.ReportAllocs()
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(zstdInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("deflate", func(b *testing.B) {
		c := NewDeflate(3, nil)
		b.ReportAllocs()
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(deflateInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("lzw", func(b *testing.B) {
		c := NewLZW(lzw.LSB, 5)
		b.ReportAllocs()
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(lzwInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("br", func(b *testing.B) {
		c := NewBrotli(brotli.WriterOptions{})
		b.ReportAllocs()
		for b.Loop() {
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
	b.ReportAllocs()
	for b.Loop() {
		w.Reset()

		wc, err := c.NewEncoder(w)
		a.NotError(err).
			NotNil(wc).
			NotError(wc.Close())
	}
}
