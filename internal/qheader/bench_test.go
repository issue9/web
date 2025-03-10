// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package qheader

import (
	"mime"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"
)

func BenchmarkParseWithParam(b *testing.B) {
	b.Run("ParseWithParam", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			v, p := ParseWithParam("application/json;charset=utf-8;k1=v1;k2=v2", "charset")
			if v != "application/json" || p != "utf-8" {
				b.Log("bench error")
			}
		}
	})

	b.Run("mime.ParseMediaType", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			mime.ParseMediaType("application/json;charset=utf-8;k1=v1;k2=v2")
		}
	})
}

func BenchmarkParseQHeader(b *testing.B) {
	a := assert.New(b, false)

	b.Run("5", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
		for b.Loop() {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
		}
	})

	b.Run("pool-5", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
		for b.Loop() {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
			PutQHeader(&items)
		}
	})

	b.Run("1", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9"
		for b.Loop() {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
		}
	})

	b.Run("pool-1", func(b *testing.B) {
		b.ResetTimer()
		str := "application/json;q=0.9"
		for b.Loop() {
			items := ParseQHeader(str, "*/*")
			a.True(len(items) > 0)
			PutQHeader(&items)
		}
	})
}

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b, false)

	for b.Loop() {
		a.Equal(BuildContentType(header.JSON, header.UTF8), "application/json; charset=utf-8")
	}
}
