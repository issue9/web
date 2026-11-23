// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation_test

import (
	"bytes"
	"compress/lzw"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/negotiation"
	"github.com/issue9/web/negotiation/compressors"
)

func TestNegotiation_ContentEncoding(t *testing.T) {
	a := assert.New(t, false)

	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddCompressor(compressors.NewLZW(lzw.LSB, 8), header.Plain, "application/*").
		AddCompressor(compressors.NewGzip(8), header.Plain).
		AddCompressor(compressors.NewZstd(), "application/*").
		AddCompressor(compressors.NewBrotli(brotli.WriterOptions{}), "text/*")

	r := &bytes.Buffer{}
	rc, err := nego.ContentEncoding("zstd", r)
	a.NotError(err).NotNil(rc)

	r = bytes.NewBufferString("123")
	rc, err = nego.ContentEncoding("", r)
	a.NotError(err).NotNil(rc)
	data, err := io.ReadAll(rc)
	a.NotError(err).Equal(string(data), "123")
}

func TestNegotiation_AcceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	nego := negotiation.New[*web.Context](language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddCompressor(compressors.NewLZW(lzw.LSB, 8), header.Plain, "application/*").
		AddCompressor(compressors.NewGzip(8), header.Plain).
		AddCompressor(compressors.NewGzip(8), "application/*")

	a.Equal(nego.AcceptEncodingHeader(), "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := a.New(t)

		b, err := nego.AcceptEncoding(header.JSON, "gzip;q=0.9,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// gzip 无法解析，br 能解析但是未添加到 Negotiation
		b, err = nego.AcceptEncoding(header.JSON, "gzip;q=x.9,br")
		a.NotError(err).Nil(b)

		b, err = nego.AcceptEncoding(header.JSON, "br,gzip")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = nego.AcceptEncoding(header.Plain, "gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = nego.AcceptEncoding(header.Plain, "br")
		a.NotError(err).Nil(b)

		b, err = nego.AcceptEncoding(header.Plain, "")
		a.NotError(err).Equal(b.Name(), "compress")
	})

	t.Run("header=*", func(t *testing.T) {
		a := a.New(t)

		b, err := nego.AcceptEncoding(header.XML, "*;q=0")
		a.NotError(err).Nil(b)

		b, err = nego.AcceptEncoding(header.XML, "*,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "compress")

		b, err = nego.AcceptEncoding(header.XML, "*,gzip")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = nego.AcceptEncoding(header.XML, "*,gzip,compress")
		a.NotError(err).Equal(b.Name(), "gzip")
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, err := nego.AcceptEncoding(header.XML, "identity,gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// 正常匹配
		b, err = nego.AcceptEncoding(header.XML, "identity;q=0,gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// identity=0
		b, err = nego.AcceptEncoding(header.XML, "identity;q=0,abc,def")
		a.Equal(err, localeutil.Error("client disabled identity")).Nil(b)
	})
}
