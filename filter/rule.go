// SPDX-License-Identifier: MIT

package filter

import (
	"fmt"
	"strconv"

	"github.com/issue9/localeutil"
)

// RulerFuncOf 数据验证规则
//
// 这是验证器与错误信息的组合。
//
// 传递参数为字段名与需要验证的值；
// 返回字段名和错误信息，如果验证成功，则返回两个空值；
type RulerFuncOf[T any] func(string, T) (string, localeutil.Stringer)

func NewRule[T any](v func(T) bool, msg localeutil.Stringer) RulerFuncOf[T] {
	return func(name string, val T) (string, localeutil.Stringer) {
		if v(val) {
			return "", nil
		}
		return name, msg
	}
}

// NewRules 将多个规则合并为一个
//
// 按顺序依次验证，直接碰到第一个验证不过的。
func NewRules[T any](r ...RulerFuncOf[T]) RulerFuncOf[T] {
	return func(name string, val T) (string, localeutil.Stringer) {
		for _, rule := range r {
			if n, msg := rule(name, val); msg != nil {
				return n, msg
			}
		}
		return "", nil
	}
}

// NewSliceRule 声明用于验证切片元素的规则
func NewSliceRule[T any, S ~[]T](v func(T) bool, msg localeutil.Stringer) RulerFuncOf[S] {
	return func(name string, val S) (string, localeutil.Stringer) {
		for index, vv := range val {
			if !v(vv) {
				return name + "[" + strconv.Itoa(index) + "]", msg
			}
		}
		return "", nil
	}
}

func NewSliceRules[T any, S ~[]T](r ...RulerFuncOf[T]) RulerFuncOf[S] {
	return func(name string, val S) (string, localeutil.Stringer) {
		for _, rule := range r {
			for index, item := range val {
				if _, msg := rule(name, item); msg != nil {
					return name + "[" + strconv.Itoa(index) + "]", msg
				}
			}
		}
		return "", nil
	}
}

// NewMapRule 声明验证 map 元素的规则
func NewMapRule[K comparable, V any, M ~map[K]V](v func(V) bool, msg localeutil.Stringer) RulerFuncOf[M] {
	return func(name string, val M) (string, localeutil.Stringer) {
		for key, vv := range val {
			if !v(vv) {
				return fmt.Sprintf("%s[%v]", name, key), msg
			}
		}
		return "", nil
	}
}

func NewMapRules[K comparable, V any, M ~map[K]V](r ...RulerFuncOf[V]) RulerFuncOf[M] {
	return func(name string, val M) (string, localeutil.Stringer) {
		for _, rule := range r {
			for key, item := range val {
				if _, msg := rule(name, item); msg != nil {
					return fmt.Sprintf("%s[%v]", name, key), msg
				}
			}
		}
		return "", nil
	}
}
