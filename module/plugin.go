// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"path/filepath"
	"plugin"
	"runtime"
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInitFuncName = "Init"

// 指定支持 plugin 模式的系统类型，需要保持该值与
// internal/plugintest/plugintest.go 中的 +build 指令中的值一致
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
func (ms *Modules) loadPlugins(glob string) error {
	if !isPluginOS() {
		return errors.New("当前平台并未实现插件功能！")
	}

	fs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, path := range fs {
		if err := ms.loadPlugin(path); err != nil {
			return err
		}
	}

	return nil
}

func (ms *Modules) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup(moduleInitFuncName)
	if err != nil {
		return err
	}
	init := symbol.(func(*Module))

	m := newModule(ms, "", "")
	init(m)

	if m.Name == "" || m.Description == "" {
		return errors.New("name 和 description 都不能为空")
	}

	// 只有在插件不出问题的情况下，才将其添加到 modules 表中
	ms.appendModules(m)

	return nil
}
