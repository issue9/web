// SPDX-License-Identifier: MIT

// Package locales 为 web 包提供了本地化的内容
package locales

import (
	"embed"

	"github.com/issue9/localeutil"
)

//go:embed *.yaml
var Locales embed.FS

// 一些多处用到的翻译项
const (
	ShouldGreatThanZero     = localeutil.StringPhrase("should great than 0")
	InvalidValue            = localeutil.StringPhrase("invalid value")
	CanNotBeEmpty           = localeutil.StringPhrase("can not be empty")
	DuplicateValue          = localeutil.StringPhrase("duplicate value")
	UniqueIdentityGenerator = localeutil.StringPhrase("unique identity generator")
)
