// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strconv"

	"github.com/issue9/context"
	"github.com/issue9/logs"
)

// ParamString 用于将路由项中的指定参数转换成字符串。
// 若该值不存在，第二个值返回false
func ParamString(w http.ResponseWriter, r *http.Request, key string) (val string, found bool) {
	m, found := context.Get(r).Get("params")
	if !found {
		logs.Debug("web.ParamString:在context中找不到params参数")
		return "", false
	}

	params := m.(map[string]string)
	val, found = params[key]
	if !found {
		logs.Debug("web.ParamString:在context.params中找不到指定参数:", key)
		return "", false
	}

	return val, true
}

// ParamInt64 功能同ParamString，但会尝试将返回值转换成int64类型。
// 若不能找到该参数，返回false
func ParamInt64(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	v, ok := ParamString(w, r, key)
	if !ok {
		return 0, false
	}

	num, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		logs.Errorf("web.ParamInt64:将参数[%v]转换成int64时，出现以下错误:%v", v, err)
		return 0, false
	}

	return num, true
}

// ParamID 功能同ParamInt64，但值必须大于0，否则第二个参数返回false。
func ParamID(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	num, ok := ParamInt64(w, r, key)
	if !ok {
		return 0, false
	}

	if num <= 0 {
		logs.Debug("ParamID:用户指定了一个小于0的id值:", num)
		RenderJSON(w, http.StatusNotFound, nil, nil)
		return 0, false
	}

	return num, true
}
