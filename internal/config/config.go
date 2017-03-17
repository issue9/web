// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 提供了框架对自身的配置文件的操作能力。
//
// 框架自身的各个模块若需要操作配置文件，应该统一交由
// Config 来管理，模块只要实现 configer 接口及一
// 个 DefaultConfig() 函数即可。
package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/server"
)

const filename = "web.json" // 配置文件的文件名。

// 需要写入到 web.json 配置文件的类需要实现的接口。
type configer interface {
	// 对配置内容进行初始化，对一些可以为空的值，进行默认赋值。
	Init() error

	// 检测配置项是否有错误。
	Check() error
}

// Config 默认的配置文件。
type Config struct {
	// 配置文件所在的目录
	dir string

	// Server
	Server *server.Config `json:"server"`

	// Content
	Content *content.Config `json:"content"`
}

// New 声明一个 *Config 实例，从 confDir/web.json 中获取。
func New(confDir string) (*Config, error) {
	conf := &Config{
		dir: confDir,
	}

	if err := conf.load(); err != nil {
		return nil, err
	}
	return conf, nil
}

// Load 加载配置文件
//
// path 用于指定配置文件的位置；
func (conf *Config) load() error {
	data, err := ioutil.ReadFile(conf.File(filename))
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, conf); err != nil {
		return err
	}

	if conf.Server == nil {
		conf.Server = server.DefaultConfig()
	}
	if err = initItem(conf.Server); err != nil {
		return err
	}

	if conf.Content == nil {
		conf.Content = content.DefaultConfig()
	}
	if err = initItem(conf.Content); err != nil {
		return err
	}

	return nil
}

func initItem(conf configer) error {
	if err := conf.Init(); err != nil {
		return err
	}

	return conf.Check()
}

// File 获取配置目录下的文件。
func (conf *Config) File(path string) string {
	return filepath.Join(conf.dir, path)
}

// DefaultConfig 输出默认配置内容。
func DefaultConfig() *Config {
	return &Config{
		Server:  server.DefaultConfig(),
		Content: content.DefaultConfig(),
	}
}
