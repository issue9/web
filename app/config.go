// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/issue9/is"
	"github.com/issue9/utils"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/web/context"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

// 端口的定义
const (
	httpPort  = 80
	httpsPort = 443
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

	OutputEncoding string `yaml:"outputEncoding"`   // 输出的编码方式
	OutputCharset  string `yaml:"outputCharset"`    // 输出的字符集名称
	Strict         bool   `yaml:"strict,omitempty"` // 参考 context.New() 中的 strict 参数

	// Domain 网站的主域名
	//
	// 必须为一个合法的域名、IP 或是 localhost 字符串。
	//
	// 当 AllowedDomain 值不会空时，此值会自动合并到 AllowedDomains 中。
	Domain string `yaml:"domain"`

	// Root 表示网站所在的根目录
	//
	// 当网站不在根目录下时，需要指定 Root，比如将网站部署在：example.com/blog
	// 则除了将 Domain 的值设置为 example.com 之外，也要将 Root 的值设置为 /blog。
	//
	// Root 值的格式必须为以 / 开头，不以 / 结尾。
	Root string `yaml:"root,omitempty"`

	HTTPS     bool              `yaml:"https"`              // 是否启用 HTTPS
	HTTPState string            `yaml:"httpState"`          // 80 端口的状态，仅在 HTTPS 为 true 时启作用
	CertFile  string            `yaml:"certFile,omitempty"` // 当 https 为 true 时，此值为必填
	KeyFile   string            `yaml:"keyFile,omitempty"`  // 当 https 为 true 时，此值为必填
	Port      int               `yaml:"port,omitempty"`     // 端口，不指定，默认为 80 或是 443
	Headers   map[string]string `yaml:"headers,omitempty"`  // 附加的头信息，头信息可能在其它地方被修改
	Options   bool              `yaml:"options,omitempty"`  // 是否启用 OPTIONS 请求
	Version   string            `yaml:"version,omitempty"`  // 限定版本

	// Static 静态内容，键名为 URL 路径，键值为文件地址
	//
	// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
	// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
	// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
	Static map[string]string `yaml:"static,omitempty"`

	// AllowedDomains 限定访问域名。
	//
	// 若指定了此值，则只有此列表与 Domain 中指定的域名可以访问当前网页。
	// 诸如 IP 和其它域名的指向将不再启作用。
	//
	// 在 AllowedDomains 中至少存在一个及以上的域名时，Domain
	// 中指定的域名会自动合并到当前列表中。
	// AllowedDomains 为空时，并不会限定域名为 Domain 指定的域名。
	AllowedDomains []string `yaml:"allowedDomains,omitempty"`

	// 指定需要加载的插件
	//
	// 每一个元素指定一条插件的路径。确保路径和权限正确。
	Plugins []string `yaml:"plugins,omitempty"`

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
	if !is.URL(conf.Domain) && conf.Domain != "localhost" {
		return errors.New("conf.domain 必须是一个 URL")
	}

	if conf.Root == "/" {
		conf.Root = conf.Root[:0]
	} else if len(conf.Root) > 0 {
		if conf.Root[0] != '/' {
			return errors.New("root 必须以 / 开始")
		}
		if conf.Root[len(conf.Root)-1] == '/' {
			return errors.New("root 不能以 / 结尾")
		}
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

	if conf.Port == 0 {
		if conf.HTTPS {
			conf.Port = httpsPort
		} else {
			conf.Port = httpPort
		}
	}

	if len(conf.AllowedDomains) > 0 {
		found := false
		for _, host := range conf.AllowedDomains {
			if !is.URL(host) {
				return fmt.Errorf("AllowedDomains 中的 %v 为非法的 URL", host)
			}

			if host == conf.Domain {
				found = true
			}
		}

		// 仅在存在 allowedDomains 字段，且该字段不为空时，才添加 domain 字段到 allowedDomains 中
		if !found {
			conf.AllowedDomains = append(conf.AllowedDomains, conf.Domain)
		}
	}

	for _, path := range conf.Plugins {
		if !utils.FileExists(path) {
			return errors.New("插件不存在")
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
