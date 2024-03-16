// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package header

import (
	"mime"
	"testing"

	"github.com/issue9/assert/v4"
)

func BenchmarkParseWithParam(b *testing.B) {
	b.Run("ParseWithParam", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			v, p := ParseWithParam("application/json;charset=utf-8;k1=v1;k2=v2", "charset")
			if v != "application/json" || p != "utf-8" {
				b.Log("bench error")
			}
		}
	})

	b.Run("mime.ParseMediaType", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			mime.ParseMediaType("application/json;charset=utf-8;k1=v1;k2=v2")
		}
	})
}

func BenchmarkParseQHeader(b *testing.B) {
	a := assert.New(b, false)

	b.Run("5", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
		for range b.N {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
		}
	})

	b.Run("pool-5", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
		for range b.N {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
			PutQHeader(&items)
		}
	})

	b.Run("1", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9"
		for range b.N {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
		}
	})

	b.Run("pool-1", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9"
		for range b.N {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
			PutQHeader(&items)
		}
	})
}

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b, false)

	for range b.N {
		a.Equal(BuildContentType("application/json", UTF8Name), "application/json; charset=utf-8")
	}
}
