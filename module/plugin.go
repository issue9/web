// SPDX-License-Identifier: MIT

package module

import (
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"runtime"
)

// 插件中的初始化函数名称，必须为可导出的函数名称
const moduleInstallFuncName = "Init"

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
func (srv *Server) loadPlugins(glob string) error {
	if !isPluginOS() {
		return errors.New("当前平台并未实现插件功能！")
	}

	fs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, path := range fs {
		if err := srv.loadPlugin(path); err != nil {
			return err
		}
	}

	return nil
}

func (srv *Server) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup(moduleInstallFuncName)
	if err != nil {
		return err
	}

	install, ok := symbol.(func(*Server))
	if !ok {
		return fmt.Errorf("插件 %s 未找到安装函数", path)
	}

	InstallFunc(install)(srv)

	return nil
}
