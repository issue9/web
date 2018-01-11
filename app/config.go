// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/issue9/is"
	"github.com/issue9/utils"
	"github.com/issue9/web/context"
	yaml "gopkg.in/yaml.v2"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

const pprofPath = "/debug/pprof/"

// 端口的定义
const (
	httpPort  = ":80"
	httpsPort = ":443"
)

// 当启用 HTTPS 时，对 80 端口的处理方式。
const (
	httpStateDisabled = "disable"  // 禁止监听 80 端口
	httpStateListen   = "listen"   // 监听 80 端口，与 HTTPS 相同的方式处理
	httpStateRedirect = "redirect" // 监听 80 端口，并重定向到 HTTPS
)

type config struct {
	// Debug 是否启用调试模式
	//
	// 该值可能会同时影响多个方面，比如是否启用 Pprof、panic 时的输出处理等
	Debug bool `yaml:"debug"`

	// Root 表示网站所在的根目录。
	//
	// 带域名，存在路径和非默认端口也得带上。
	// 若带路径，则在构建路由时，会自动加此前缀。
	Root string `yaml:"root"`

	OutputEncoding string `yaml:"outputEncoding"`   // 输出的编码方式
	OutputCharset  string `yaml:"outputCharset"`    // 输出的字符集名称
	Strict         bool   `yaml:"strict,omitempty"` // 参考 context.New() 中的 strict 参数

	HTTPS     bool              `yaml:"https"`              // 是否启用 HTTPS
	HTTPState string            `yaml:"httpState"`          // 80 端口的状态，仅在 HTTPS 为 true 时启作用
	CertFile  string            `yaml:"certFile,omitempty"` // 当 https 为 true 时，此值为必填
	KeyFile   string            `yaml:"keyFile,omitempty"`  // 当 https 为 true 时，此值为必填
	Port      string            `yaml:"port,omitempty"`     // 端口，不指定，默认为 80 或是 443
	Headers   map[string]string `yaml:"headers,omitempty"`  // 附加的头信息，头信息可能在其它地方被修改
	Static    map[string]string `yaml:"static,omitempty"`   // 静态内容，键名为 URL 路径，键值为文件地址
	Options   bool              `yaml:"options,omitempty"`  // 是否启用 OPTIONS 请求
	Version   string            `yaml:"version,omitempty"`  // 限定版本
	Hosts     []string          `yaml:"hosts,omitempty"`    // 限定访问域名。仅需指定域名

	// 性能
	ReadTimeout  time.Duration `yaml:"readTimeout"`  // http.Server.ReadTimeout 的值
	WriteTimeout time.Duration `yaml:"writeTimeout"` // http.Server.WriteTimeout 的值
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

	if conf.HTTPS {
		switch conf.HTTPState {
		case httpStateListen, httpStateDisabled, httpStateRedirect:
		default:
			return errors.New("httpState 的值不正确")
		}

		if !utils.FileExists(conf.CertFile) {
			return errors.New("certFile 文件不存在")
		}

		if !utils.FileExists(conf.KeyFile) {
			return errors.New("keyFile 文件不存在")
		}
	}

	if len(conf.Port) == 0 {
		if conf.HTTPS {
			conf.Port = httpsPort
		} else {
			conf.Port = httpPort
		}
	} else if conf.Port[0] != ':' {
		conf.Port = ":" + conf.Port
	}

	if len(conf.Hosts) > 0 {
		for _, host := range conf.Hosts {
			if !is.URL(host) {
				return fmt.Errorf("Hosts 中的 %v 为非法的 URL", host)
			}
		}
	}

	if conf.ReadTimeout < 0 {
		return errors.New("readTimeout 必须大于等于 0")
	}

	if conf.WriteTimeout < 0 {
		return errors.New("writeTimeout 必须大于等于 0")
	}
	return nil
}
