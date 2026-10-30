// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import "github.com/issue9/localeutil"

// Collector 错误信息收集函数
type Collector func(name string, reason localeutil.Stringer)

// Filter 过滤器
//
// 包含了数据验证和数据修正功能
type Filter struct {
	name        string
	exitOnError bool
	collector   Collector
	len         int
}

// New 声明过滤器
//
// name 过滤器的名称，如果不这空，那么所有验证属性名称都会加上此值；
func New(name string, exitOnError bool, collecter Collector) *Filter {
	return &Filter{
		name:        name,
		exitOnError: exitOnError,
		collector:   collecter,
	}
}

// Name 由 [New] 的参数 name
func (v *Filter) Name() string { return v.name }

func (v *Filter) ExitOnError() bool { return v.exitOnError }

// AddReason 直接添加一条错误信息
func (v *Filter) AddReason(name string, reason localeutil.Stringer) *Filter {
	if v.len > 0 && v.exitOnError {
		return v
	}

	if v.name != "" {
		name = v.name + name
	}

	v.len++
	v.collector(name, reason)
	return v
}

// AddError 添加一条类型为 err 的错误信息
func (v *Filter) AddError(name string, err error) *Filter {
	if ls, ok := err.(localeutil.Stringer); ok {
		return v.AddReason(name, ls)
	}
	return v.AddReason(name, localeutil.Phrase(err.Error()))
}

// Add 使用 rule 规则验证字段 name 表示的值 value
func (v *Filter) Add[T any](name string, value *T, rule ...Rule[T]) *Filter {
	if v.len > 0 && v.exitOnError {
		return v
	}

	for _, r := range rule {
		if name, ls := r(name, value); ls != nil {
			v.AddReason(name, ls)
			break
		}
	}

	return v
}

func (v *Filter) AddFilter(name string, f func(f *Filter)) *Filter {
	f(New(v.name+name, v.exitOnError, v.collector))
	return v
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *Filter) When(cond bool, f func(v *Filter)) *Filter {
	if cond {
		f(v)
	}
	return v
}
