// SPDX-License-Identifier: MIT

// Package locales 为 web 包提供了本地化的内容
//
// 并不主动加载这些信息，用户可以根据自身的需求引用 [Locales]。
package locales

import "embed"

//go:embed *.yml
var Locales embed.FS
