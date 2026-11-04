// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package filter 提供数据验证和修正功能
package filter

import "github.com/issue9/localeutil"

// Filter 过滤器
//
// 包含了数据验证和数据修正功能
//
// 过滤器中的各类验证和修改函数都是立即执行的，如果存在依赖关系，需要注意调用顺序。
type Filter struct {
	name      string
	collector Collector

	exitOnError bool
	hasError    bool
}

// New 声明过滤器
//
// name 过滤器的名称，如果不为空，那么所有验证属性名称都会加上此值；
func New(name string, exitOnError bool, collector Collector) *Filter {
	if collector == nil {
		panic("参数 collector 不能为空")
	}

	return &Filter{
		name:        name,
		exitOnError: exitOnError,
		collector:   collector,
	}
}

// Reset 重置对象
func (f *Filter) Reset(name string, exitOnError bool, collector Collector) {
	if collector == nil {
		panic("参数 collector 不能为空")
	}

	f.name = name
	f.exitOnError = exitOnError
	f.collector = collector
	f.hasError = false
}

// Name 由 [New] 的参数 name
func (f *Filter) Name() string { return f.name }

func (f *Filter) ExitOnError() bool { return f.exitOnError }

func (f *Filter) HasError() bool { return f.hasError }

// AddReason 直接添加一条错误信息
//
// name 表示 reason 关联的名称，如果为空，则会使用 [Filter.Name] 代替，否则使用 `Filter.Name() + '.' + name` 组合。
func (f *Filter) AddReason(name string, reason localeutil.Stringer) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}
	return f.addReason(name, reason)
}

func (f *Filter) addReason(name string, reason localeutil.Stringer) *Filter {
	if reason == nil {
		panic("参数 name 和 reason 不能为空")
	}

	if f.name != "" {
		if name == "" {
			name = f.name
		} else {
			name = f.name + "." + name
		}
	}

	if name == "" {
		panic("参数 name 无法获取正确的值")
	}

	f.hasError = true
	f.collector(name, reason)
	return f
}

// AddError 添加一条类型为 err 的错误信息
//
// name 表示 err 关联的名称，如果为空，则会使用 [Filter.Name] 代替，否则使用 `Filter.Name() + '.' + name` 组合。
func (f *Filter) AddError(name string, err error) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	if ls, ok := err.(localeutil.Stringer); ok {
		return f.addReason(name, ls)
	}
	return f.addReason(name, localeutil.Phrase(err.Error()))
}

// ApplyFilter 执行一个过滤器函数
//
// 该过滤器产生的错误信息字段名会以参数 name 作为前缀。
//
// name 表示该过滤器函数产生的错误信息关联名称，
// 如果为空，则会使用 [Filter.Name] 代替，否则使用 `Filter.Name() + '.' + name` 组合。
func (f *Filter) ApplyFilter(name string, fun func(f *Filter)) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	if name == "" {
		if f.Name() == "" {
			panic("参数 name 无法获取正确的值")
		}

		fun(f)
		return f
	}

	if f.name != "" {
		name = f.name + "." + name
	}

	ff := New(name, f.ExitOnError(), f.collector)

	fun(ff)
	if ff.HasError() {
		f.hasError = true
	}

	return f
}

// When 只有满足 cond 才执行 f 中的验证
//
// fun 的参数即为当前对象；
func (f *Filter) When(cond func() bool, fun func(*Filter)) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	if cond() {
		fun(f)
	}
	return f
}
