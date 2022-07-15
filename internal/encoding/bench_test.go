// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/issue9/assert/v2"
)

func BenchmarkEncodingWriterBuilder_Build(b *testing.B) {
	a := assert.New(b, false)

	e := NewEncodings(nil, "text*")
	a.NotNil(e)
	a.False(e.allowAny).
		Empty(e.ignoreTypes).
		Equal(e.ignoreTypePrefix, []string{"text"})
	e.Add(map[string]WriterFunc{
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
	}
}
