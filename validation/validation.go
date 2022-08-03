// SPDX-License-Identifier: MIT

// Package validation 验证功能
package validation

import (
	"reflect"
	"strconv"
	"sync"

	"github.com/issue9/localeutil"
)

const validationPoolMaxSize = 20

var validationPool = &sync.Pool{New: func() any {
	return &Validation{
		keys:    make([]string, 0, 5),
		reasons: make([]localeutil.LocaleStringer, 0, 5),
	}
}}

var (
	isNotSlice = localeutil.Phrase("the type is not slice or array")
	isNotMap   = localeutil.Phrase("the type is not map")
)

// Validation 验证工具
type Validation struct {
	exitAtError bool

	keys    []string
	reasons []localeutil.LocaleStringer
}

// New 声明验证对象
//
// exitAtError 表示是在验证出错时，是否还继续其它字段的验证。
func New(exitAtError bool) *Validation {
	v := validationPool.Get().(*Validation)
	v.exitAtError = exitAtError
	v.keys = v.keys[:0]
	v.reasons = v.reasons[:0]
	return v
}

func (v *Validation) Count() int { return len(v.keys) }

// Visit 依次访问每一条错误信息
//
// f 为访问错误信息的方法，其原型为：
//  func(key string, reason localeutil.LocaleStringer) (ok bool)
// 其中的 key 为名称，reason 为出错原因，返回值 ok 表示是否继续下一条信息的访问。
func (v *Validation) Visit(f func(string, localeutil.LocaleStringer) bool) {
	for index, key := range v.keys {
		if !f(key, v.reasons[index]) {
			break
		}
	}
}

// Destroy 回收当前对象
//
// 这不是一个必须调用的方法，但是该方法在大量频繁地的使用 Validation 时有一定的性能提升。
func (v *Validation) Destroy() {
	if v.Count() < validationPoolMaxSize {
		validationPool.Put(v)
	}
}

// Add 直接添加一条错误信息
//
// 此方法不受 exitAtError 标记位的影响。
func (v *Validation) Add(name string, reason localeutil.LocaleStringer) {
	v.keys = append(v.keys, name)
	v.reasons = append(v.reasons, reason)
}

// AddField 验证新的字段
//
// val 表示需要被验证的值；
// name 表示当前字段的名称，当验证出错时，以此值作为名称返回给用户；
// rules 表示验证的规则，按顺序依次验证。
func (v *Validation) AddField(val any, name string, rules ...*Rule) *Validation {
	if v.Count() > 0 && v.exitAtError {
		return v
	}

	for _, rule := range rules {
		if !rule.validator.IsValid(val) {
			v.Add(name, rule.message)
			break
		}
	}
	return v
}

// AddSliceField 验证数组字段
//
// 如果字段类型不是数组或是字符串，将添加一条错误信息，并退出验证。
func (v *Validation) AddSliceField(val any, name string, rules ...*Rule) *Validation {
	// TODO: 如果 go 支持泛型方法，那么可以将 val 固定在 []T

	if v.Count() > 0 && v.exitAtError {
		return v
	}

	rv := reflect.ValueOf(val)

	if kind := rv.Kind(); kind != reflect.Array && kind != reflect.Slice && kind != reflect.String {
		v.Add(name, isNotSlice)
		return v
	}

	for i := 0; i < rv.Len(); i++ {
		for _, rule := range rules {
			if !rule.validator.IsValid(rv.Index(i).Interface()) {
				v.Add(name+"["+strconv.Itoa(i)+"]", rule.message)
				if v.exitAtError {
					return v
				}
			}
		}
	}

	return v
}

// AddMapField 验证 map 字段
//
// 如果字段类型不是 map，将添加一条错误信息，并退出验证。
func (v *Validation) AddMapField(val any, name string, rules ...*Rule) *Validation {
	// TODO: 如果 go 支持泛型方法，那么可以将 val 固定在 map[T]T

	if v.Count() > 0 && v.exitAtError {
		return v
	}

	rv := reflect.ValueOf(val)
	if kind := rv.Kind(); kind != reflect.Map {
		v.Add(name, isNotMap)
		return v
	}

	keys := rv.MapKeys()
	for i := 0; i < rv.Len(); i++ {
		key := keys[i]
		for _, rule := range rules {
			if !rule.validator.IsValid(rv.MapIndex(key).Interface()) {
				v.Add(name+"["+key.String()+"]", rule.message)
				if v.exitAtError {
					return v
				}
			}
		}
	}

	return v
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *Validation) When(cond bool, f func(v *Validation)) *Validation {
	if cond {
		f(v)
	}
	return v
}
