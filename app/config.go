// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/issue9/is"
	"github.com/issue9/utils"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/server"
	yaml "gopkg.in/yaml.v2"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

type config struct {
	// 表示网站的根目录，带域名，非默认端口也得带上。
	Root string `yaml:"root"`

	// 是否启用调试模式
	Debug bool `yaml:"debug"`

	// 输出的编码方式，必须已经通过 context.AddUnmarshal() 添加
	OutputEncoding string `yaml:"outputEncoding"`

	// 输出的字符符，必须已经通过 context.AddCharset() 添加
	OutputCharset string `yaml:"outputCharset"`

	// 参考 context.New() 中的 strict 参数
	Strict bool `yaml:"strict"`

	// Server
	Server *server.Config `yaml:"server"`
}

// 加载配置文件
//
// path 用于指定配置文件的位置；
func loadConfig(path string) (*config, error) {
	conf := &config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}

	if err = conf.sanitize(); err != nil {
		return nil, err
	}

	return conf, nil
}

// 修正可修正的内容，返回不可修正的错误。
func (conf *config) sanitize() error {
	if !is.URL(conf.Root) || strings.Contains(conf.Root, "localhost:") {
		return errors.New("conf.Root 必须是一个 URL")
	}
	if strings.HasSuffix(conf.Root, "/") {
		conf.Root = conf.Root[:len(conf.Root)-1]
	}

	if conf.OutputCharset == "" {
		conf.OutputCharset = context.DefaultCharset
	}

	if conf.OutputEncoding == "" {
		conf.OutputEncoding = context.DefaultEncoding
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

	return nil
}
