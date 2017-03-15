// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	stdCtx "context"
	"errors"
	"path/filepath"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/utils"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/config"
	"github.com/issue9/web/modules"
	"github.com/issue9/web/result"
)

// Version 当前框架的版本
const Version = "0.5.0+20170315"

const (
	configFilename = "web.json" // 配置文件的文件名。
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
)

var (
	confDir        string         // 配置文件所在目录
	defaultConfig  *config.Config // 当前的配置实例
	defaultContent content.Content
	defaultModules = modules.New() // 模块管理工具
)

// Init 初始化框架的基本内容。
//
// configDir 指定了配置文件所在的目录，框架默认的
// 两个配置文件都会从此目录下查找。
func Init(configDir string) error {
	if !utils.FileExists(configDir) {
		return errors.New("配置文件目录不存在")
	}
	confDir = configDir

	return load()
}

// 加载配置，初始化相关的组件。
func load() error {
	// 初始化日志系统，第一个初始化，后续内容可能都依赖于此。
	err := logs.InitFromXMLFile(File(logsFilename))
	if err != nil {
		return err
	}

	// 加载配置文件
	defaultConfig, err = config.Load(File(configFilename))
	if err != nil {
		return err
	}

	// 确定编码
	e := defaultConfig.Envelope
	state := content.EnvelopeStateDisable
	switch e.State {
	case config.EnvelopeStateDisable:
		state = content.EnvelopeStateDisable
	case config.EnvelopeStateEnable:
		state = content.EnvelopeStateEnable
	case config.EnvelopeStateMust:
		state = content.EnvelopeStateMust
	}
	defaultContent, err = content.New(defaultConfig.ContentType, state, e.Key, e.Status)
	if err != nil {
		return err
	}

	return nil
}

// Run 运行路由，执行监听程序。
func Run() error {
	if err := defaultModules.Init(); err != nil {
		return err
	}

	return run(defaultConfig, defaultServeMux)
}

// Restart 重启服务。
//
// timeout 等待该时间之后重启，小于该值，则立即重启。
func Restart(timeout time.Duration) error {
	if err := Shutdown(timeout); err != nil {
		return err
	}

	// 重新加载配置内容
	if err := load(); err != nil {
		return err
	}

	return Run()
}

// Shutdown 关闭服务。
//
// timeout 若超过该时间，服务还未自动停止的，则会强制停止。
// 若 timeout<=0，则会立即停止服务，相当于 http.Server.Close()；
// 若 timeout>0 时，则会等待处理完毕或是该时间耗尽才停止服务，相当于 http.Server.Shutdown()。
func Shutdown(timeout time.Duration) error {
	if timeout <= 0 {
		for _, srv := range servers {
			if err := srv.Close(); err != nil {
				return err
			}
		}
		clearServers()

		return nil
	}

	ctx, cancel := stdCtx.WithTimeout(stdCtx.Background(), timeout)
	defer cancel()

	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
	}
	clearServers()

	return nil
}

// File 获取配置目录下的文件。
func File(path string) string {
	return filepath.Join(confDir, path)
}

// IsDebug 是否处于调试状态。
// 系统通过判断是否指定了 pprof 配置项，来确定当前是否处于调试状态。
func IsDebug() bool {
	return len(defaultConfig.Pprof) > 0
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
