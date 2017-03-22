// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/utils"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/config"
	"github.com/issue9/web/internal/server"
	"github.com/issue9/web/modules"
)

// Version 当前框架的版本
const Version = "0.8.1+20170322"

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.json" // 配置文件的文件名。
)

var (
	configDir      string         // 配置文件所在目录
	defaultConfig  *config.Config // 当前的配置实例
	defaultServer  *server.Server
	defaultContent content.Content
	defaultModules = modules.New() // 模块管理工具
)

// Init 初始化框架的基本内容。
//
// confDir 指定了配置文件所在的目录，框架默认的
// 两个配置文件都会从此目录下查找。
func Init(confDir string) error {
	if !utils.FileExists(confDir) {
		return errors.New("配置文件目录不存在")
	}
	configDir = confDir

	return load()
}

// 加载配置，初始化相关的组件。
func load() error {
	err := logs.InitFromXMLFile(File(logsFilename))
	if err != nil {
		return err
	}

	defaultConfig, err = config.Load(File(configFilename))
	if err != nil {
		return err
	}

	defaultServer, err = server.New(defaultConfig.Server)
	if err != nil {
		return err
	}

	defaultContent, err = content.New(defaultConfig.Content)
	if err != nil {
		return err
	}

	return nil
}

// Run 运行路由，执行监听程序。
func Run(h http.Handler) error {
	if err := defaultModules.Init(); err != nil {
		return err
	}

	defaultServer.Init(h)

	return defaultServer.Run()
}

// Restart 重启路由服务。仅是重启路由服务，
// 不会重新加载配置文件，不然会影响全局内容，和直接重启程序没什么区别。
//
// timeout 等待该时间之后重启，若小于或等于 0 则立即重启。
func Restart(timeout time.Duration) error {
	logs.All("重启服务器...")
	logs.Flush()

	if defaultServer != nil {
		return defaultServer.Restart(timeout)
	}

	return nil
}

// Shutdown 关闭所有服务。
// 关闭之后不能再调用 Run() 重新运行。
// 若只是想重启服务，只能调用 Restart() 函数。
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func Shutdown(timeout time.Duration) error {
	logs.Flush()

	if defaultServer != nil {
		if err := defaultServer.Shutdown(timeout); err != nil {
			return err
		}
	}

	defaultConfig = nil
	defaultContent = nil
	defaultServer = nil
	defaultModules = nil

	return nil
}

// File 获取配置目录下的文件。
func File(path string) string {
	return filepath.Join(configDir, path)
}

// NewModule 注册一个新的模块。
//
// name 为模块名称；
// init 当前模块的初始化函数；
// deps 模块的依赖模块，这些模块在初始化时，会先于 name 初始化始。
func NewModule(name string, init modules.Init, deps ...string) {
	err := defaultModules.New(name, init, deps...)

	// 注册模块时出错，直接退出。
	if err != nil {
		logs.Fatal(err)
	}
}
