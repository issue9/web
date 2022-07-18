// SPDX-License-Identifier: MIT

package serialization

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v2"
)

func BenchmarkMimetypes_MarshalFunc(b *testing.B) {
	a := assert.New(b, false)
	mt := NewMimetypes(10)
	a.NotNil(mt)

	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		name, marshal, ok := mt.MarshalFunc("font/wottf;q=0.9")
		a.True(ok).NotEmpty(name).NotNil(marshal)
	}
}

func BenchmarkMimetypes_UnmarshalFunc(b *testing.B) {
	a := assert.New(b, false)
	mt := NewMimetypes(10)
	a.NotNil(mt)

	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		marshal, ok := mt.UnmarshalFunc("font/wottf")
		a.True(ok).NotNil(marshal)
	}
}
