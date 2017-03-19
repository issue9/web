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

	"github.com/issue9/utils"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/server"
)

const filename = "web.json" // 配置文件的文件名。

// 需要写入到 web.json 配置文件的类需要实现的接口。
type sanitizer interface {
	// 修正可修正的内容，返回不可修正的错误。
	Sanitize() error
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

	// Server
	if conf.Server == nil {
		conf.Server = server.DefaultConfig()
	} else {
		// 将未设置的项，给予一个默认值。
		c := server.DefaultConfig()
		if err = utils.Merge(true, c, conf.Server); err != nil {
			return err
		}
		conf.Server = c

		if err = conf.Server.Sanitize(); err != nil {
			return err
		}
	}

	// Content
	if conf.Content == nil {
		conf.Content = content.DefaultConfig()
	} else {
		c := content.DefaultConfig()
		if err = utils.Merge(true, c, conf.Content); err != nil {
			return err
		}
		conf.Content = c

		if err = conf.Content.Sanitize(); err != nil {
			return err
		}
	}

	return nil
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
