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

	"github.com/issue9/web/content"
	"github.com/issue9/web/server"
)

// Config 系统配置文件。
type Config struct {
	Server *server.Config `json:"server"`

	// Content
	Content *content.Config `json:"content,omitempty"`
}

// Load 加载配置文件
//
// path 用于指定配置文件的位置；
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	if err = json.Unmarshal(data, conf); err != nil {
		return nil, err
	}

	return conf, nil
}
