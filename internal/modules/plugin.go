// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package modules

import (
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"runtime"

	"github.com/issue9/web/module"
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInitFuncName = "Init"

// 指定支持 plugin 模式的系统类型，需要保持该值与 gen.go 中的 +build 指令中的值一致
var pluginOS = []string{"linux", "darwin"}

func isPluginOS() bool {
	for _, os := range pluginOS {
		if os == runtime.GOOS {
			return true
		}
	}

	return false
}

// 加载所有的插件
//
// 如果 glob 为空，则不会加载任何内容，返回空值
func loadPlugins(glob string) ([]*module.Module, error) {
	if !isPluginOS() {
		return nil, errors.New("windows 平台并未实现插件功能！")
	}

	fs, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	modules := make([]*module.Module, 0, len(fs))
	for _, path := range fs {
		m, err := loadPlugin(path)
		if err != nil {
			return nil, err
		}

		modules = append(modules, m)
	}

	return modules, nil
}

func loadPlugin(path string) (*module.Module, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	symbol, err := p.Lookup(moduleInitFuncName)
	if err != nil {
		return nil, err
	}
	init := symbol.(func(*module.Module))

	m := module.New(module.TypePlugin, "", "plugin desc")
	init(m)

	if m.Name == "" {
		return nil, fmt.Errorf("插件 %s 未指定插件名称", path)
	}

	return m, nil
}
