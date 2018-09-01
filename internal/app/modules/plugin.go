// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"fmt"
	"path/filepath"
	"plugin"

	"github.com/issue9/mux"
	"github.com/issue9/web/module"
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInitFuncName = "Init"

// 加载所有的插件
//
// 如果 glob 为空，则不会加载任何内容，返回空值
func loadPlugins(glob string, router *mux.Prefix) ([]*module.Module, error) {
	fs, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	modules := make([]*module.Module, 0, len(fs))
	for _, path := range fs {
		m, err := loadPlugin(path, router)
		if err != nil {
			return nil, err
		}

		modules = append(modules, m)
	}

	return modules, nil
}

func loadPlugin(path string, router *mux.Prefix) (*module.Module, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	symbol, err := p.Lookup(moduleInitFuncName)
	if err != nil {
		return nil, err
	}
	init := symbol.(func(*module.Module))

	m := module.New("", "plugin desc")
	m.Type = module.TypePlugin
	init(m)

	if m.Name == "" {
		return nil, fmt.Errorf("插件 %s 未指定插件名称", path)
	}

	return m, nil
}
