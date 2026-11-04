// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package locales 为 web 包提供了本地化的内容
package locales

import (
	"cmp"
	"embed"
	"io/fs"

	cache "github.com/issue9/cache/locales"
	config "github.com/issue9/config/locales"
	"github.com/issue9/localeutil"
	messages "github.com/issue9/messages/locales"
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
	messages.Locales,
	config.Locales,
	cache.Locales,
	scheduled.Locales,
}

//---------------------- 一些多处复用的本地化对象会出现在这里 ---------------------------

const (
	NotFound                = localeutil.StringPhrase("not found")
	InvalidFormat           = localeutil.StringPhrase("invalid format")
	InvalidValue            = localeutil.StringPhrase("invalid value")
	CanNotBeEmpty           = localeutil.StringPhrase("can not be empty")
	DuplicateValue          = localeutil.StringPhrase("duplicate value")
	UniqueIdentityGenerator = localeutil.StringPhrase("unique identity generator")
)

// NewFieldError 返回表示配置文件错误的对象
//
// field 表示错误的字段名；
func NewFieldError(field string, msg any) error {
	return localeutil.Error("field %s has error: %s", field, msg)
}

// ShouldGreatThan 返回必须大于 n 的翻译项
func ShouldGreatThan[T cmp.Ordered](n T) localeutil.Stringer {
	return localeutil.Phrase("should great than %v", n)
}

var (
	errInvalidValue  = localeutil.Error("invalid value")
	errCanNotBeEmpty = localeutil.Error("can not be empty")
)

func ErrInvalidValue() error { return errInvalidValue }

func ErrCanNotBeEmpty() error { return errCanNotBeEmpty }
