// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"github.com/issue9/web/result"
)

// NewResult 生成一个 *result.Result 对象
func NewResult(code int, fields map[string]string) *result.Result {
	return result.New(code, fields)
}
