// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation_test

import (
	"bytes"
	"compress/lzw"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/negotiation"
	"github.com/issue9/web/negotiation/codecs"
	"github.com/issue9/web/negotiation/compressors"
)

func BenchmarkNegotiation_Accept(b *testing.B) {
	a := assert.New(b, false)
	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	nego.AddCodec(codecs.NewJSON(), header.JSON, "application/problem+json", "", negotiation.PatchKindNone).
		AddCodec(codecs.NewXML(), header.XML, "application/problem+xml", "", negotiation.PatchKindNone).
		AddCodec(codecs.NewYAML(), "application/test", "application/problem+test", "", negotiation.PatchKindNone)

	b.ReportAllocs()
	var item *negotiation.MediaType[*web.Context]
	for b.Loop() {
		item, _ = nego.Accept("application/json;q=0.9")
	}
	a.NotNil(item)
}

func BenchmarkNegotiation_ContentType(b *testing.B) {
	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	nego.AddCodec(codecs.NewJSON(), header.JSON, "application/problem+json", "", negotiation.PatchKindNone).
		AddCodec(codecs.NewXML(), header.XML, "application/problem+xml", "", negotiation.PatchKindNone).
		AddCodec(codecs.NewYAML(), "application/test", "application/problem+test", "", negotiation.PatchKindNone)

	b.Run("charset=utf-8", func(b *testing.B) {
		a := assert.New(b, false)
		var (
			mt       *negotiation.MediaType[*web.Context]
			encoding encoding.Encoding
			err      error
		)

		b.ReportAllocs()
		for b.Loop() {
			mt, encoding, err = nego.ContentType(qheader.BuildContentType(header.XML, header.UTF8), "")
		}
		a.NotError(err).NotNil(mt).Nil(encoding)
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()
		for b.Loop() {
			marshal, encoding, err := nego.ContentType(qheader.BuildContentType(header.XML, "gbk"), "")
			a.NotError(err).NotNil(marshal).NotNil(encoding)
		}
	})
}

func BenchmarkNegotiation_ContentEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
		a.NotNil(c)
		c.AddCompressor(compressors.NewZstd(), "application/*")
		b.ReportAllocs()

		r := bytes.NewBuffer([]byte{})
		var err error
		for b.Loop() {
			r.Reset()
			_, err = c.ContentEncoding("zstd", r)
		}
		a.NotError(err)
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()

		c := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
		a.NotNil(c)
		c.AddCompressor(compressors.NewGzip(8), "application/*").
			AddCompressor(compressors.NewBrotli(brotli.WriterOptions{}), "text/*").
			AddCompressor(compressors.NewDeflate(8, nil), "image/*").
			AddCompressor(compressors.NewZstd(), "application/*").
			AddCompressor(compressors.NewLZW(lzw.LSB, 8), header.Plain)

		r := bytes.NewBuffer([]byte{})
		var err error
		for b.Loop() {
			r.Reset()
			_, err = c.ContentEncoding("zstd", r)
		}
		a.NotError(err)
	})
}

func BenchmarkNegotiation_AcceptEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()

		c := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
		a.NotNil(c)
		c.AddCompressor(compressors.NewZstd(), "application/*")

		var err error
		for b.Loop() {
			_, err = c.AcceptEncoding(header.JSON, "zstd")
		}
		a.NotError(err)
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()

		c := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
		a.NotNil(c)
		c.AddCompressor(compressors.NewGzip(8), "application/*").
			AddCompressor(compressors.NewBrotli(brotli.WriterOptions{}), "text/*").
			AddCompressor(compressors.NewDeflate(8, nil), "image/*").
			AddCompressor(compressors.NewZstd(), "application/*").
			AddCompressor(compressors.NewLZW(lzw.LSB, 8), header.Plain)

		var err error
		for b.Loop() {
			_, err = c.AcceptEncoding(header.Plain, "compress")
		}
		a.NotError(err)
	})
}
