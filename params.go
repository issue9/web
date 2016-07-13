// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strconv"

	"github.com/issue9/context"
)

// ParamString 获取一个 string 类型的参数，
// 若不存在，则第二个参数返回 false，并向 logs.DEBUG() 输出一条信息。
func ParamString(r *http.Request, key string) (string, bool) {
	m, found := context.Get(r).Get("params")
	if !found {
		Debug(r, "web.ParamString:在context中找不到params参数")
		return "", false
	}

	params := m.(map[string]string)
	val, found := params[key]
	if !found {
		Debug(r, "web.ParamString:在context.params中找不到指定参数:", key)
		return "", false
	}

	return val, true
}

// ParamInt64 获取一个 int64 类型的参数，
// 若不存在或转换出错，则第二个参数返回 false，并向相应的日志通道输出一条信息。
func ParamInt64(r *http.Request, key string) (int64, bool) {
	str, ok := ParamString(r, key)
	if !ok {
		return 0, false
	}

	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		Errorf(r, "web.ParamInt64:将参数[%v]转换成int64时，出现以下错误:%v", str, err)
		return 0, false
	}

	return num, true
}

// ParamInt 获取一个 int 类型的参数，
// 若不存在或转换出错，则第二个参数返回 false，并向相应的日志通道输出一条信息。
func ParamInt(r *http.Request, key string) (int, bool) {
	str, ok := ParamString(r, key)
	if !ok {
		return 0, false
	}

	num, err := strconv.Atoi(str)
	if err != nil {
		Errorf(r, "web.ParamInt:将参数[%v]转换成int64时，出现以下错误:%v", str, err)
		return 0, false
	}

	return num, true
}

// ParamFloat64 获取一个 float64 类型的参数，
// 若不存在或转换出错，则第二个参数返回 false，并向相应的日志通道输出一条信息。
func ParamFloat64(r *http.Request, key string) (float64, bool) {
	str, ok := ParamString(r, key)
	if !ok {
		return 0, false
	}

	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		Errorf(r, "web.ParamFloat64:将参数[%v]转换成float64时，出现以下错误:%v", str, err)
		return 0, false
	}

	return num, true
}

// ParamID 获取一个大于 0 的 int64 类型的参数，
// 若不存在或转换出错，则第二个参数返回 false，并向相应的日志通道输出一条信息。
func ParamID(r *http.Request, key string) (int64, bool) {
	num, ok := ParamInt64(r, key)
	if !ok {
		return 0, false
	}

	if num <= 0 {
		Debug(r, "web.ParamID:用户指定了一个小于0的id值:", num)
		return 0, false
	}

	return num, true
}
