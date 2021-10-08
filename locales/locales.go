// SPDX-License-Identifier: MIT

// Package locales 提供当前包本地化的信息
package locales

import "embed"

//go:embed *.yml
var Locales embed.FS
