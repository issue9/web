// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 提供了程序对自身的配置文件的操作能力。
//
// NOTE: 所有需要写入到配置文件的配置项，都应该在此定义。
package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/issue9/web/content"
	"github.com/issue9/web/server"
)

const filename = "web.json" // 配置文件的文件名。

// Config 系统配置文件。
type Config struct {
	confDir string // 配置文件所在的目录

	// Server
	Server *server.Config `json:"server"`

	// Content
	Content *content.Config `json:"content,omitempty"`
}

// New 声明一个 *Config 实例
func New(confDir string) (*Config, error) {
	conf := &Config{
		confDir: confDir,
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

	return nil
}

// File 获取配置目录下的文件。
func (conf *Config) File(path string) string {
	return filepath.Join(conf.confDir, path)
}
