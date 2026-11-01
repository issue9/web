// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"errors"
	"testing"

	"github.com/issue9/localeutil"
)

var benchFilter *Filter

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	c := newCollector()

	for b.Loop() {
		benchFilter = New("", false, c.collect)
	}
}

func BenchmarkFilter_Reset(b *testing.B) {
	b.ReportAllocs()
	c := newCollector()
	f := New("reset", true, c.collect)

	for b.Loop() {
		f.Reset("reset", false, c.collect)
	}
}

func BenchmarkFilter_AddFilter(b *testing.B) {
	b.ReportAllocs()
	f, _ := newFilter(false)

	for b.Loop() {
		f.AddFilter("v1.", func(f *Filter) {
			f.AddReason("f1", localeutil.StringPhrase("s1"))
			f.AddFilter("v2.", func(f *Filter) {
				f.AddError("f2", errors.New("s2"))
			})
		})
	}
}
