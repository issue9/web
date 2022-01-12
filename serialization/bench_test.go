// SPDX-License-Identifier: MIT

package serialization

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"io"
	"log"
	"testing"

	"github.com/issue9/assert/v2"
)

func BenchmarkMimetypes_MarshalFunc(b *testing.B) {
	a := assert.New(b, false)
	srv := NewMimetypes(10)
	a.NotNil(srv)

	a.NotError(srv.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		name, marshal, ok := srv.MarshalFunc("font/wottf;q=0.9")
		a.True(ok).NotEmpty(name).NotNil(marshal)
	}
}

func BenchmarkMimetypes_UnmarshalFunc(b *testing.B) {
	a := assert.New(b, false)
	srv := NewMimetypes(10)
	a.NotNil(srv)

	a.NotError(srv.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		marshal, ok := srv.UnmarshalFunc("font/wottf")
		a.True(ok).NotNil(marshal)
	}
}

func BenchmarkEncodingWriterBuilder_Build(b *testing.B) {
	a := assert.New(b, false)

	e := NewEncodings(log.Default(), "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})
	e.Add(map[string]EncodingWriterFunc{
		"gzip": gzipWriterFunc,
		"br":   gzipWriterFunc,
	})

	builder, notAccept := e.Search("application/json", "gzip;q=0.9,br")
	a.False(notAccept).NotNil(builder).Equal(builder.name, "br")

	for i := 0; i < b.N; i++ {
		w := &bytes.Buffer{}
		wc := builder.Build(w)
		_, err := wc.Write([]byte("123456"))
		a.NotError(err)
		a.NotError(wc.Close())

		r, err := gzip.NewReader(w)
		a.NotError(err).NotNil(r)
		data, err := io.ReadAll(r)
		a.NotError(err).NotNil(data).Equal(string(data), "123456")

		builder.Put(wc)
	}
}
