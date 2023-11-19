// SPDX-License-Identifier: MIT

package locale

import (
	"testing"

	"golang.org/x/text/language"
)

func BenchmarkLocale_NewPrinter(b *testing.B) {
	l := New(language.SimplifiedChinese, nil, nil)

	b.Run("equal id", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.NewPrinter(language.SimplifiedChinese)
		}
	})

	b.Run("not equal id", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.NewPrinter(language.TraditionalChinese)
		}
	})
}
