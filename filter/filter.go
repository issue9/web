// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package filter 提供数据验证和修正功能
package filter

import "github.com/issue9/localeutil"

// Collector 错误信息收集函数
type Collector func(name string, reason localeutil.Stringer)

// Filter 过滤器
//
// 包含了数据验证和数据修正功能
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
func (f *Filter) AddReason(name string, reason localeutil.Stringer) *Filter {
	if name == "" || reason == nil {
		panic("参数 name 和 reason 不能为空")
	}

	if f.HasError() && f.ExitOnError() {
		return f
	}

	if name == "" {
		name = f.name
	} else if f.name != "" {
		name = f.name + "." + name
	}

	f.hasError = true
	f.collector(name, reason)
	return f
}

// AddError 添加一条类型为 err 的错误信息
func (f *Filter) AddError(name string, err error) *Filter {
	if ls, ok := err.(localeutil.Stringer); ok {
		return f.AddReason(name, ls)
	}
	return f.AddReason(name, localeutil.Phrase(err.Error()))
}

// Add 使用 rule 规则验证字段 name 表示的值 value
func (f *Filter) Add[T any](name string, value *T, rule ...Rule[T]) *Filter {
	if value == nil {
		panic("参数 value 不能为空")
	}

	if f.HasError() && f.ExitOnError() {
		return f
	}

	for _, r := range rule {
		if name, ls := r(name, value); ls != nil {
			f.AddReason(name, ls)
			break
		}
	}

	return f
}

// AddFilter 执行一个过滤器函数
//
// 该过滤器产生的错误信息字段名会以参数 name 作为前缀。
// 如果 name 为空，那么直接传递当前对象 f 给 fun 函数。
func (f *Filter) AddFilter(name string, fun func(f *Filter)) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	if name == "" {
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
// f 中的 v 即为当前对象；
func (f *Filter) When(cond func() bool, fun func(v *Filter)) *Filter {
	if f.HasError() && f.ExitOnError() {
		return f
	}

	if cond() {
		fun(f)
	}
	return f
}
