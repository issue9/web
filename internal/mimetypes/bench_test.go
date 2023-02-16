// SPDX-License-Identifier: MIT

package mimetypes

import (
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v3"
)

func BenchmarkMimetypes_MarshalFunc(b *testing.B) {
	a := assert.New(b, false)
	mt := New[marshalFunc, unmarshalFunc](10)
	a.NotNil(mt)

	mt.Add("font/wottf", xml.Marshal, xml.Unmarshal, "")

	for i := 0; i < b.N; i++ {
		item := mt.MarshalFunc("font/wottf;q=0.9")
		a.NotNil(item)
	}
}

func BenchmarkMimetypes_ContentType(b *testing.B) {
	a := assert.New(b, false)
	mt := New[marshalFunc, unmarshalFunc](10)
	a.NotNil(mt)

	mt.Add("font/1", xml.Marshal, xml.Unmarshal, "")
	mt.Add("font/2", xml.Marshal, xml.Unmarshal, "")
	mt.Add("font/3", xml.Marshal, xml.Unmarshal, "")

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
