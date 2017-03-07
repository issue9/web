// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/utils"
	"github.com/issue9/web/contentype"
	"github.com/issue9/web/internal/config"
)

// Version 当前框架的版本
const Version = "0.2.1+20170307"

const (
	configFilename = "web.json" // 配置文件的文件名。
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
)

var (
	confDir            string                  // 配置文件所在目录
	defaultConfig      *config.Config          // 当前的配置实例
	defaultContentType contentype.ContentTyper // 编码解码工具
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
	defaultContentType, err = contentype.New(defaultConfig.ContentType, logs.ERROR())
	return err
}

// File 获取相对于配置目录的文件路径。
func File(path string) string {
	return filepath.Join(confDir, path)
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
// timeout 等待该时间之后重启。
func Restart(timeout time.Duration) error {
	if err := Shutdown(timeout); err != nil {
		return err
	}

	return Run()
}

// Shutdown 关闭服务。
//
// 若 timeout<=0，则会立即停止服务，相当于 http.Server.Close()；
// 若 timeout>0 时，则会等待处理完毕或是该时间耗尽才停止服务，相当于 http.Server.Shutdown()。
func Shutdown(timeout time.Duration) error {
	if timeout <= 0 {
		for _, srv := range servers {
			if err := srv.Close(); err != nil {
				return err
			}
		}

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

// IsDebug 是否处于调试状态。
// 系统通过判断是否指定了 pprof 配置项，来确定当前是否处于调试状态。
func IsDebug() bool {
	return len(defaultConfig.Pprof) > 0
}
