// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package locale

import (
	"testing"

	"golang.org/x/text/language"
)

func BenchmarkLocale_NewPrinter(b *testing.B) {
	l := New(language.SimplifiedChinese, nil, nil)

	b.Run("equal Locale.id", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.NewPrinter(language.SimplifiedChinese)
		}
	})

	b.Run("not equal Locale.id", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.NewPrinter(language.TraditionalChinese)
		}
	})

	langs := []language.Tag{
		language.Chinese,
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.MustParse("zh-CN"),
		language.MustParse("cmn-Hans"),
	}
	b.Run("rand id", func(b *testing.B) {
		b.ResetTimer()
		size := len(langs)
		for i := 0; i < b.N; i++ {
			l.NewPrinter(langs[i%size])
		}
	})
}
