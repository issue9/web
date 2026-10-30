// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"errors"
	"testing"

	"github.com/issue9/localeutil"
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	c := newCollector()

	for b.Loop() {
		_ = New("", false, c.collect)
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
