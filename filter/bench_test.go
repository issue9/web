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
	_, c := CreateMapCollector()

	for b.Loop() {
		benchFilter = New("", false, c)
	}
}

func BenchmarkFilter_Reset(b *testing.B) {
	b.ReportAllocs()
	_, c := CreateMapCollector()
	f := New("reset", true, c)

	for b.Loop() {
		f.Reset("reset", false, c)
	}
}

func BenchmarkFilter_AddFilter(b *testing.B) {
	b.ReportAllocs()
	f, _ := newFilter(false, "")

	for b.Loop() {
		f.ApplyFilter("v1.", func(f *Filter) {
			f.AddReason("f1", localeutil.StringPhrase("s1"))
			f.ApplyFilter("v2.", func(f *Filter) {
				f.AddError("f2", errors.New("s2"))
			})
		})
	}
}
