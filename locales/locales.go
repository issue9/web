// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package locales 为 web 包提供了本地化的内容
package locales

import (
	"embed"
	"io/fs"

	cache "github.com/issue9/cache/locales"
	config "github.com/issue9/config/locales"
	"github.com/issue9/localeutil"
	localeutilL "github.com/issue9/localeutil/locales"
	scheduled "github.com/issue9/scheduled/locales"
)

//go:embed *.yaml
var locales embed.FS

// Locales 当前框架依赖的所有本地化内容
//
// 文件格式均为 yaml，使用时加载这些文件系统下的 yaml 文件即可：
//
//	s := server.New(...)
//	s.Locale().LoadMessages("*.yaml", locales.Locales()...)
var Locales = []fs.FS{
	locales,
	localeutilL.Locales,
	config.Locales,
	cache.Locales,
	scheduled.Locales,
}

//---------------------------- 以下为本地化的文本内容 -----------------------------

const (
	NotFound                = localeutil.StringPhrase("not found")
	InvalidFormat           = localeutil.StringPhrase("invalid format")
	InvalidValue            = localeutil.StringPhrase("invalid value")
	CanNotBeEmpty           = localeutil.StringPhrase("can not be empty")
	DuplicateValue          = localeutil.StringPhrase("duplicate value")
	UniqueIdentityGenerator = localeutil.StringPhrase("unique identity generator")
	RecycleLocalCache       = localeutil.StringPhrase("recycle local cache")
)

// ShouldGreatThan 返回必须大于 n 的翻译项
func ShouldGreatThan[T any](n T) localeutil.Stringer {
	return localeutil.Phrase("should great than %v", n) // n 可以是时间等类型
}

//---------------------------- 以下为本地化的错误实例 -----------------------------

var (
	errInvalidFormat = localeutil.Error("invalid format")
	errInvalidValue  = localeutil.Error("invalid value")
	errCanNotBeEmpty = localeutil.Error("can not be empty")
	errNotFound      = localeutil.Error("not found")
)

func ErrInvalidFormat() error { return errInvalidFormat }

func ErrInvalidValue() error { return errInvalidValue }

func ErrCanNotBeEmpty() error { return errCanNotBeEmpty }

func ErrNotFound() error { return errNotFound }
