// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import "github.com/issue9/web/result"

// NewResult 声明一个新的 *result.Result 实例
func NewResult(code int) *result.Result {
	return result.New(code)
}

// NewResultWithDetail 声明一个新的 *result.Result 实例
func NewResultWithDetail(code int, detail map[string]string) *result.Result {
	return result.NewWithDetail(code, detail)
}

// NewMessage 注册一条新的信息
func NewMessage(code int, message string) error {
	return result.NewMessage(code, message)
}

// NewMessages 批量注册信息
func NewMessages(messages map[int]string) error {
	return result.NewMessages(messages)
}
