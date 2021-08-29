// SPDX-License-Identifier: MIT

package config

import "github.com/issue9/localeutil"

var (
	_ error                     = &Error{}
	_ localeutil.LocaleStringer = &Error{}
)
