// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"strconv"

	"github.com/issue9/web"
)

// Operation 提供根据的 [Operation] 生成 openapi 文档的中间件
func (d *Document) Operation(o *Operation) web.Middleware {
	return web.MiddlewareFunc(func(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
		d.addOperation(method, pattern, router, o)
		return next
	})
}

func (api *Document) addOperation(method, pattern, _ string, opt *Operation) {
	if api.paths == nil {
		api.paths = make(map[string]*PathItem, 50)
	}

	opt.addComponents(api.components)

	for _, ref := range api.headers {
		opt.Headers = append(opt.Headers, &Parameter{
			Ref: &Ref{Ref: ref},
		})
	}
	for _, ref := range api.cookies {
		opt.Cookies = append(opt.Cookies, &Parameter{
			Ref: &Ref{Ref: ref},
		})
	}
	for status, ref := range api.responses {
		code := strconv.Itoa(status)
		if _, found := opt.Responses[code]; !found {
			opt.Responses[code] = &Response{Ref: &Ref{Ref: ref}}
		}
	}

	if item, found := api.paths[pattern]; !found {
		item = &PathItem{Operations: make(map[string]*Operation, 3)}
		item.Operations[method] = opt
		api.paths[pattern] = item
	} else {
		if _, found = item.Operations[method]; found {
			panic(fmt.Sprintf("已经存在 %s:%s 的定义", method, pattern))
		}
		item.Operations[method] = opt
	}
}
