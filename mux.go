// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/mux"
)

var defaultServeMux = mux.NewServeMux()

// Clean 清除所有的路由项
func Clean() *mux.ServeMux {
	return defaultServeMux.Clean()
}

// Remove 移除指定的路由项，通过路由表达式和 method 来匹配。
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func Remove(pattern string, methods ...string) {
	defaultServeMux.Remove(pattern, methods...)
}

// Add 添加一条路由数据。
func Add(pattern string, h http.Handler, methods ...string) *mux.ServeMux {
	return defaultServeMux.Add(pattern, h, methods...)
}

// Options 手动指定 OPTIONS 请求方法的值。
func Options(pattern string, allowMethods ...string) *mux.ServeMux {
	return defaultServeMux.Options(pattern, allowMethods...)
}

// Get 相当于 defaultServeMux.Add(pattern, h, "GET") 的简易写法
func Get(pattern string, h http.Handler) *mux.ServeMux {
	return defaultServeMux.Get(pattern, h)
}

// Post 相当于 defaultServeMux.Add(pattern, h, "POST") 的简易写法
func Post(pattern string, h http.Handler) *mux.ServeMux {
	return defaultServeMux.Post(pattern, h)
}

// Delete 相当于 defaultServeMux.Add(pattern, h, "DELETE") 的简易写法
func Delete(pattern string, h http.Handler) *mux.ServeMux {
	return defaultServeMux.Delete(pattern, h)
}

// Put 相当于 defaultServeMux.Add(pattern, h, "PUT") 的简易写法
func Put(pattern string, h http.Handler) *mux.ServeMux {
	return defaultServeMux.Put(pattern, h)
}

// Patch 相当于 defaultServeMux.Add(pattern, h, "PATCH") 的简易写法
func Patch(pattern string, h http.Handler) *mux.ServeMux {
	return defaultServeMux.Patch(pattern, h)
}

// Any 相当于 defaultServeMux.Add(pattern, h) 的简易写法
func Any(pattern string, h http.Handler) *mux.ServeMux {
	return defaultServeMux.Any(pattern, h)
}

// AddFunc 相当于 defaultServeMux.AddFunc(pattern, func, ...) 的简易写法
func AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *mux.ServeMux {
	return defaultServeMux.AddFunc(pattern, fun, methods...)
}

// GetFunc 相当于 defaultServeMux.AddFunc(pattern, func, "GET") 的简易写法
func GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *mux.ServeMux {
	return defaultServeMux.GetFunc(pattern, fun)
}

// PutFunc 相当于 defaultServeMux.AddFunc(pattern, func, "PUT") 的简易写法
func PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *mux.ServeMux {
	return defaultServeMux.PutFunc(pattern, fun)
}

// PostFunc 相当于 defaultServeMux.AddFunc(pattern, func, "POST") 的简易写法
func PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *mux.ServeMux {
	return defaultServeMux.PostFunc(pattern, fun)
}

// DeleteFunc 相当于 defaultServeMux.AddFunc(pattern, func, "DELETE") 的简易写法
func DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *mux.ServeMux {
	return defaultServeMux.DeleteFunc(pattern, fun)
}

// PatchFunc 相当于 defaultServeMux.AddFunc(pattern, func, "PATCH") 的简易写法
func PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *mux.ServeMux {
	return defaultServeMux.PatchFunc(pattern, fun)
}

// AnyFunc 相当于 defaultServeMux.AddFunc(pattern, func) 的简易写法
func AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *mux.ServeMux {
	return defaultServeMux.AnyFunc(pattern, fun)
}
