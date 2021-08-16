// SPDX-License-Identifier: MIT

package serialization

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkMimetypes_MarshalFunc(b *testing.B) {
	a := assert.New(b)
	srv := NewMimetypes(10)
	a.NotNil(srv)

	a.NotError(srv.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		name, marshal, err := srv.MarshalFunc("font/wottf;q=0.9")
		a.NotError(err).
			NotEmpty(name).
			NotNil(marshal)
	}
}

func BenchmarkMimetypes_UnmarshalFunc(b *testing.B) {
	a := assert.New(b)
	srv := NewMimetypes(10)
	a.NotNil(srv)

	a.NotError(srv.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		marshal, err := srv.UnmarshalFunc("font/wottf")
		a.NotError(err).
			NotNil(marshal)
	}
}
