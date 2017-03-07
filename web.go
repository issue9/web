// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"path/filepath"

	"github.com/issue9/logs"
	"github.com/issue9/utils"
	"github.com/issue9/web/contentype"
	"github.com/issue9/web/internal/config"
)

// Version 当前框架的版本
const Version = "0.1.0+20170307"

const (
	configFilename = "web.json" // 配置文件的文件名。
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
)

var (
	confDir       string         // 配置文件所在目录
	defaultConfig *config.Config // 当前的配置实例
)

// Init 初始化系统的基础架构。
// 包括加载配置文件、初始化日志系统、数据库等。
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

// File 获取相对于配置目录的文件路径
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

// IsDebug 是否处于调试状态
func IsDebug() bool {
	return len(defaultConfig.Pprof) > 0
}
