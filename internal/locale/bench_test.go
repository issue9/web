// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package locale

import (
	"testing"

	"golang.org/x/text/language"
)

func BenchmarkLocale_NewPrinter(b *testing.B) {
	l := New(language.SimplifiedChinese, nil)

	b.Run("equal Locale.id", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			l.NewPrinter(language.SimplifiedChinese)
		}
	})

	b.Run("not equal Locale.id", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
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
		for i := range b.N {
			l.NewPrinter(langs[i%size])
		}
	})
}
