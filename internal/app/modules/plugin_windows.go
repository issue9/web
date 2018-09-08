// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"errors"

	"github.com/issue9/mux"

	"github.com/issue9/web/module"
)

// 加载所有的插件
//
// 如果 glob 为空，则不会加载任何内容，返回空值
func loadPlugins(glob string, router *mux.Prefix) ([]*module.Module, error) {
	return nil, errors.New("windows 平台并未实现插件功能！")
}
