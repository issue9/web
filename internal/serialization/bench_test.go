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

func BenchmarkMimetypes_unmarshalFunc(b *testing.B) {
	a := assert.New(b, false)
	mt := NewMimetypes(10)
	a.NotNil(mt)

	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, "font/wottf"))

	for i := 0; i < b.N; i++ {
		marshal, ok := mt.unmarshalFunc("font/wottf")
		a.True(ok).NotNil(marshal)
	}
}

func BenchmarkMimetypes_ContentType(b *testing.B) {
	a := assert.New(b, false)
	mt := NewMimetypes(10)
	a.NotNil(mt)

	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, "font/1"))
	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, "font/2"))
	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, "font/3"))

	b.Run("charset=utf-8", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			marshal, encoding, err := mt.ContentType("font/2;charset=utf-8")
			a.NotError(err).NotNil(marshal).Nil(encoding)
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			marshal, encoding, err := mt.ContentType("font/2;charset=gbk")
			a.NotError(err).NotNil(marshal).NotNil(encoding)
		}
	})
}
