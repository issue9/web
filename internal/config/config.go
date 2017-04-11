// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 提供了框架对自身的配置文件的操作能力。
//
// 框架自身的各个模块若需要操作配置文件，应该统一交由
// Config 来管理，模块只要实现 sanitizer 接口及一
// 个 DefaultConfig() 函数即可。
package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/issue9/is"
	"github.com/issue9/utils"
	"github.com/issue9/web/content"
	"github.com/issue9/web/internal/server"
)

// 需要写入到 web.json 配置文件的类需要实现的接口。
type sanitizer interface {
	// 修正可修正的内容，返回不可修正的错误。
	Sanitize() error
}

// Config 配置文件。
type Config struct {
	// 表示网站的根目录，带域名，非默认端口也得带上。
	Root string `json:"root"`

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

	if err = conf.Sanitize(); err != nil {
		return nil, err
	}

	return conf, nil
}

// Sanitize 修正可修正的内容，返回不可修正的错误。
func (conf *Config) Sanitize() error {
	if !is.URL(conf.Root) {
		return errors.New("conf.Root 必须是一个 URL")
	}
	if conf.Root[len(conf.Root)-1] == '/' {
		conf.Root = conf.Root[:len(conf.Root)-1]
	}

	// Server
	if conf.Server == nil {
		conf.Server = server.DefaultConfig()
	} else {
		// 将未设置的项，给予一个默认值。
		c := server.DefaultConfig()
		if err := utils.Merge(true, c, conf.Server); err != nil {
			return err
		}
		conf.Server = c

		if err := conf.Server.Sanitize(); err != nil {
			return err
		}
	}

	// Content
	if conf.Content == nil {
		conf.Content = content.DefaultConfig()
	} else {
		c := content.DefaultConfig()
		if err := utils.Merge(true, c, conf.Content); err != nil {
			return err
		}
		conf.Content = c

		if err := conf.Content.Sanitize(); err != nil {
			return err
		}
	}

	return nil
}

// DefaultConfig 输出默认配置内容。
func DefaultConfig() *Config {
	return &Config{
		Server:  server.DefaultConfig(),
		Content: content.DefaultConfig(),
	}
}
