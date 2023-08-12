// SPDX-License-Identifier: MIT

// Package locales 本地化内容
package locales

import "embed"

//go:embed *.yaml
var Locales embed.FS
