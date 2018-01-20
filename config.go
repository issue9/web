// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/issue9/is"
	"github.com/issue9/utils"

	"github.com/issue9/web/encoding"
)

type config struct {
	// Debug 是否启用调试模式
	//
	// 该值可能会同时影响多个方面，比如是否启用 Pprof、panic 时的输出处理等
	Debug bool `yaml:"debug"`

	// Strict 严格模式。
	//
	// 启用此配置，某些内容的验证会更加严格。
	// 比如会检测客户端的 Accept 是否接受当前的 OutputEncoding 值等。
	Strict bool `yaml:"strict,omitempty"`

	// OutputEncoding 向客户输出时，采用的编码方式，值类型应该采用 mime-type 值。
	//
	// 此编码方式必须已经通过 context.AddMarsal() 添加。
	//
	// 如果为空，则会采用 context.DefaultEncoding 作为默认值。
	OutputEncoding string `yaml:"outputEncoding"`

	// OutputCharset 向客户端输出的字符集名称。
	//
	// 此编码方式必须已经通过 context.AddCharset() 添加。
	//
	// 如果为空，则会采用 context.DefaultCharset 作为默认值。
	OutputCharset string `yaml:"outputCharset"`

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

	// HTTPS 是否启用 HTTPS 协议
	//
	// 如果启用此配置，则需要保证 CertFile 和 KeyFile 两个文件必须存在，
	// 这两个文件最终会被传递给 http.ListenAndServeTLS() 的两个参数。
	//
	// 此值还会影响 Port 的默认值。
	HTTPS    bool   `yaml:"https"`
	CertFile string `yaml:"certFile,omitempty"`
	KeyFile  string `yaml:"keyFile,omitempty"`
	Port     int    `yaml:"port,omitempty"`

	// DisableOptions 是否禁用对 OPTIONS 请求的处理
	DisableOptions bool `yaml:"disableOptions,omitempty"`

	// Headers 附加的报头信息
	//
	// 一些诸如跨域等报头信息，可以在此作设置。
	//
	// 报头信息可能在其它处理器被修改。
	Headers map[string]string `yaml:"headers,omitempty"`

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

	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`

	// URL 网站的首页地址。
	// 一般情况下，如果用到诸如生成 URL 地址什么的，会用到此值。
	//
	// 若为空，则会根据配置文件的内容，生成网站首页地址。
	// 用户也可台强制指定一个不同的地址。
	URL string `yaml:"url"`
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
		conf.OutputCharset = encoding.DefaultCharset
	}

	if conf.OutputEncoding == "" {
		conf.OutputEncoding = encoding.DefaultEncoding
	}

	if conf.HTTPS {
		if !utils.FileExists(conf.CertFile) {
			return errors.New("certFile 文件不存在")
		}

		if !utils.FileExists(conf.KeyFile) {
			return errors.New("keyFile 文件不存在")
		}
	}

	if conf.Port == 0 {
		if conf.HTTPS {
			conf.Port = 443
		} else {
			conf.Port = 80
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

	if conf.URL == "" {
		if conf.HTTPS {
			conf.URL = "https://" + conf.Domain
			if conf.Port != 443 {
				conf.URL += ":" + strconv.Itoa(conf.Port)
			}
		} else {
			conf.URL = "http://" + conf.Domain
			if conf.Port != 80 {
				conf.URL += ":" + strconv.Itoa(conf.Port)
			}
		}
		conf.URL += conf.Root
	}

	return nil
}
