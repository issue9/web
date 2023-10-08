// SPDX-License-Identifier: MIT

// Package locales 为 web 包提供了本地化的内容
package locales

import (
	"embed"
	"io/fs"

	cl "github.com/issue9/config/locales"
	"github.com/issue9/localeutil"
	ll "github.com/issue9/localeutil/locales"
)

//go:embed *.yaml
var Locales embed.FS

var all = []fs.FS{
	Locales,
	ll.Locales,
	cl.Locales,
}

// 一些多处用到的翻译项
const (
	ShouldGreatThanZero     = localeutil.StringPhrase("should great than 0")
	InvalidValue            = localeutil.StringPhrase("invalid value")
	CanNotBeEmpty           = localeutil.StringPhrase("can not be empty")
	DuplicateValue          = localeutil.StringPhrase("duplicate value")
	UniqueIdentityGenerator = localeutil.StringPhrase("unique identity generator")
	RecycleLocalCache       = localeutil.StringPhrase("recycle local cache")
)

// All 当前框架依赖的所有本地化内容
//
// 文件格式均为 yaml，使用时加载它些文件系统下的 yaml 文件即可。
func All() []fs.FS { return all }
