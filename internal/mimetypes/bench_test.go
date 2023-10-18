// SPDX-License-Identifier: MIT

package mimetypes

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/serializer/xml"
)

func BenchmarkMimetypes_Accept(b *testing.B) {
	a := assert.New(b, false)
	mt := New(10)
	a.NotNil(mt)

	mt.Add("font/wottf", xml.BuildMarshal, xml.Unmarshal, "")
	mt.Add("text/plain", json.BuildMarshal, json.Unmarshal, "text/plain+problem")

	for i := 0; i < b.N; i++ {
		item := mt.Accept("font/wottf;q=0.9")
		a.NotNil(item)
	}
}

func BenchmarkMimetypes_ContentType(b *testing.B) {
	a := assert.New(b, false)
	mt := New(10)
	a.NotNil(mt)

	mt.Add("font/1", xml.BuildMarshal, xml.Unmarshal, "")
	mt.Add("font/2", xml.BuildMarshal, xml.Unmarshal, "")
	mt.Add("font/3", xml.BuildMarshal, xml.Unmarshal, "")

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
