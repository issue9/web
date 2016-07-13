// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strconv"
)

// QueryString 获取查询参数 key 的值。
//
// 若该值为空或是不存在，则返回 def 作为其默认值
func QueryString(r *http.Request, key, def string) string {
	val := r.FormValue(key)
	if len(val) == 0 {
		return def
	}
	return val
}

// QueryInt 获取查询参数 key 的 int 类型值。
//
// 若 key 不存在则使用 def 作为默认参数返回。
// 在无法转换的情况下，也将 def 作为默认值返回且第二个参数还将返回 false。
func QueryInt(r *http.Request, key string, def int) (int, bool) {
	val := r.FormValue(key)
	if len(val) == 0 { // 未传递该参数值，当作使用默认值。
		return def, true
	}

	ret, err := strconv.Atoi(val)
	if err != nil {
		Errorf(r, "web.QueryInt:将查询参数[%v]转换成int时，出现以下错误:%v", val, err)
		return def, false
	}
	return ret, true
}

// QueryInt64 获取查询参数 key 的 int64 类型值。
//
// 若 key 不存在则使用 def 作为默认参数返回。
// 在无法转换的情况下，也将 def 作为默认值返回且第二个参数还将返回 false。
func QueryInt64(r *http.Request, key string, def int64) (int64, bool) {
	val := r.FormValue(key)
	if len(val) == 0 {
		return def, true
	}

	ret, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		Errorf(r, "web.QueryInt64:将查询参数[%v]转换成int64时，出现以下错误:%v", val, err)
		return def, false
	}
	return ret, true
}

// QueryFloat64 获取查询参数 key 的 float64 类型值。
//
// 若 key 不存在则使用 def 作为默认参数返回。
// 在无法转换的情况下，也将 def 作为默认值返回且第二个参数还将返回 false。
func QueryFloat64(r *http.Request, key string, def float64) (float64, bool) {
	val := r.FormValue(key)
	if len(val) == 0 {
		return def, true
	}

	ret, err := strconv.ParseFloat(val, 64)
	if err != nil {
		Errorf(r, "web.QueryFloat64:将查询参数[%v]转换成float64时，出现以下错误:%v", val, err)
		return def, false
	}
	return ret, true
}

// QueryBool 获取查询参数 key 的 bool 类型值。
//
// 若 key 不存在则使用 def 作为默认参数返回。
// 在无法转换的情况下，也将 def 作为默认值返回且第二个参数还将返回 false。
func QueryBool(r *http.Request, key string, def bool) (bool, bool) {
	val := r.FormValue(key)
	if len(val) == 0 {
		return def, true
	}

	ret, err := strconv.ParseBool(val)
	if err != nil {
		Errorf(r, "web.QueryBool:将查询参数[%v]转换成bool时，出现以下错误:%v", val, err)
		return def, false
	}

	return ret, true
}
