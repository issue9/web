// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package sanitizers

import (
	"html"
	"strings"
	"unicode"
)

// Trim 过滤左右空格
func Trim(v *string) { *v = strings.TrimSpace(*v) }

func TrimLeft(v *string) {
	*v = strings.TrimLeftFunc(*v, func(r rune) bool { return unicode.IsSpace(r) })
}

func TrimRight(v *string) {
	*v = strings.TrimRightFunc(*v, func(r rune) bool { return unicode.IsSpace(r) })
}

func Upper(v *string) { *v = strings.ToUpper(*v) }

func Lower(v *string) { *v = strings.ToLower(*v) }

func EscapeHTML(v *string) { *v = html.EscapeString(*v) }

func UnescapeHTML(v *string) { *v = html.UnescapeString(*v) }
