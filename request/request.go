// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package request 提供了一些常用的与请求有关的操作。
package request

import (
	"net/http"
	"strings"
)

// ResultFields 从报头中获取 X-Result-Fields 的相关内容。
//
// allow 表示所有允许出现的字段名称。
// 当第二个参数返回 true 时，返回的是可获取的字段名列表；
// 当第二个参数返回 false 时，返回的是不允许获取的字段名。
func ResultFields(r *http.Request, allow []string) ([]string, bool) {
	resultFields := r.Header.Get("X-Result-Fields")
	if len(resultFields) == 0 { // 没有指定，则返回所有字段内容
		return allow, true
	}
	fields := strings.Split(resultFields, ",")
	fails := make([]string, 0, len(fields))

	isAllow := func(field string) bool {
		for _, f1 := range allow {
			if f1 == field {
				return true
			}
		}
		return false
	}

	for index, field := range fields {
		field = strings.TrimSpace(field)
		fields[index] = field

		if !isAllow(field) { // 记录不允许获取的字段名
			fails = append(fails, field)
		}
	}

	if len(fails) > 0 {
		return fails, false
	}

	return fields, true
}
