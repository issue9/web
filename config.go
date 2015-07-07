// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// 配置文件所在的目录。
var configDir string

var cfg = &config{}

type config struct {
	Https      bool                 `json:"https"`            // 是否启用https
	CertFile   string               `json:"certFile"`         // 当https为true时，此值为必须
	KeyFile    string               `json:"keyFile"`          // 当https为true时，此值为必须
	Port       string               `json:"port"`             // 端口，不指定，默认为80或是443
	ServerName string               `json:"serverName"`       // 响应头的server变量，为空时，不输出该内容
	DB         map[string]*dbConfig `json:"db"`               // 数据库相关配置
	Static     map[string]string    `json:"static,omitempty"` // 表态路由映身，键名表示路由路径，键值表示文件目录
}

type dbConfig struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
	Prefix string `json:"prefix"`
}

// 从dir参数初始化configDir变量。会给路径最后加上/分隔符。
func initConfigDir(dir string) {
	// 确保dir参数以/结尾
	last := dir[len(dir)-1]
	if last != filepath.Separator && last != '/' {
		dir += string(filepath.Separator)
	}

	// 经过上面的代码处理，此时dir最后个字符为/，
	// 所以不需要再判断其是否为目录，若不是目录，err会描述其具体信息。
	if _, err := os.Stat(dir); err != nil {
		panic(err)
	}

	configDir = dir
}

// 返回config目录下的文件
func ConfigFile(filename string) string {
	return configDir + filename
}

// 从path中加载配置内容到cfg变量中
func loadConfig(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(data, cfg); err != nil {
		panic(err)
	}

	// Port检测
	if len(cfg.Port) == 0 {
		if cfg.Https {
			cfg.Port = ":443"
		} else {
			cfg.Port = ":80"
		}
	}
	if cfg.Port[0] != ':' {
		cfg.Port = ":" + cfg.Port
	}

	// 确保每个目录都以/结尾
	for k, v := range cfg.Static {
		last := v[len(v)-1]
		if last != filepath.Separator && last != '/' {
			cfg.Static[k] = v + string(filepath.Separator)
		}
	}
}
