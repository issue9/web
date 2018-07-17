// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/issue9/logs"
	yaml "gopkg.in/yaml.v2"
)

const logsFilename = "logs.xml" // 日志配置文件的文件名。

// Sanitizer 检测配置文件内容
type Sanitizer interface {
	Sanitize() error
}

// Config 配置文件管理
type Config struct {
	dir string
}

// New 声明新的 *Config 实例
func New(dir string) (*Config, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	conf := &Config{dir: dir}

	if err = logs.InitFromXMLFile(conf.File(logsFilename)); err != nil {
		return nil, err
	}

	return conf, nil
}

// Load 加载指定的配置文件内容到 v 中
func (conf *Config) Load(path string, v interface{}) error {
	ext := strings.ToLower(filepath.Ext(path))

	var unmarshal func([]byte, interface{}) error

	switch ext {
	case ".yaml", ".yml":
		unmarshal = yaml.Unmarshal
	case ".json":
		unmarshal = json.Unmarshal
	case ".xml":
		unmarshal = xml.Unmarshal
	default:
		return fmt.Errorf("无效的配置文件类型：%s", ext)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = unmarshal(data, v); err != nil {
		return err
	}

	if s, ok := v.(Sanitizer); ok {
		return s.Sanitize()
	}
	return nil
}

// File 获取配置目录下的文件名
func (conf *Config) File(path ...string) string {
	paths := make([]string, 0, len(path)+1)
	paths = append(paths, conf.dir)
	return filepath.Join(append(paths, path...)...)
}
