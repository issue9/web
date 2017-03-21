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

	"github.com/issue9/utils"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/server"
)

// 需要写入到 web.json 配置文件的类需要实现的接口。
type sanitizer interface {
	// 修正可修正的内容，返回不可修正的错误。
	Sanitize() error
}

// Config 默认的配置文件。
type Config struct {
	// Server
	Server *server.Config `json:"server"`

	// Content
	Content *content.Config `json:"content"`
}

// Load 加载配置文件
//
// path 用于指定配置文件的位置；
func Load(path string) (*Config, error) {
	conf := &Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, conf); err != nil {
		return nil, err
	}

	// Server
	if conf.Server == nil {
		conf.Server = server.DefaultConfig()
	} else {
		// 将未设置的项，给予一个默认值。
		c := server.DefaultConfig()
		if err = utils.Merge(true, c, conf.Server); err != nil {
			return nil, err
		}
		conf.Server = c

		if err = conf.Server.Sanitize(); err != nil {
			return nil, err
		}
	}

	// Content
	if conf.Content == nil {
		conf.Content = content.DefaultConfig()
	} else {
		c := content.DefaultConfig()
		if err = utils.Merge(true, c, conf.Content); err != nil {
			return nil, err
		}
		conf.Content = c

		if err = conf.Content.Sanitize(); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

// DefaultConfig 输出默认配置内容。
func DefaultConfig() *Config {
	return &Config{
		Server:  server.DefaultConfig(),
		Content: content.DefaultConfig(),
	}
}
