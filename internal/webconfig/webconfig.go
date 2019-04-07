// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package webconfig web.yaml 配置文件对应的内容。
package webconfig

import (
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/is"
	"github.com/issue9/utils"
)

const localhostURL = "localhost"

// WebConfig 项目的配置内容
type WebConfig struct {
	// Domain 网站的主域名
	//
	// 必须为一个合法的域名、IP 或是 localhost 字符串。
	//
	// 当 AllowedDomain 值不为空时，此值会自动合并到 AllowedDomains 中。
	Domain string `yaml:"domain,omitempty" json:"domain.omitempty"`

	// Debug 是否启用调试模式
	//
	// 该值可能会同时影响多个方面，比如是否启用 Pprof、panic 时的输出处理等
	Debug bool `yaml:"debug,omitempty" json:"debug,omitempty"`

	// Root 表示网站所在的根目录
	//
	// 当网站不在根目录下时，需要指定 Root，比如将网站部署在：example.com/blog
	// 则除了将 Domain 的值设置为 example.com 之外，也要将 Root 的值设置为 /blog。
	//
	// Root 值的格式必须为以 / 开头，不以 / 结尾，或是空值。
	Root string `yaml:"root,omitempty" json:"root,omitempty"`

	// Plugins 指定插件，通过 glob 语法指定，比如：~/plugins/*.so
	// 为空表示没有插件。
	//
	// 当前仅支持部分系统：https://golang.org/pkg/plugin/
	Plugins string `yaml:"plugins,omitempty" json:"plugins,omitempty"`

	// HTTPS 是否启用 HTTPS 协议
	//
	// 如果启用此配置，则需要保证 CertFile 和 KeyFile 两个文件必须存在，
	// 这两个文件最终会被传递给 http.ListenAndServeTLS() 的两个参数。
	//
	// 此值还会影响 Port 的默认值。
	HTTPS    bool   `yaml:"https,omitempty" json:"https,omitempty"`
	CertFile string `yaml:"certFile,omitempty" json:"certFile,omitempty"`
	KeyFile  string `yaml:"keyFile,omitempty" json:"keyFile,omitempty"`
	Port     int    `yaml:"port,omitempty" json:"port,omitempty"`

	// DisableOptions 是否禁用自动生成 OPTIONS 和 HEAD 请求的处理
	DisableOptions bool `yaml:"disableOptions,omitempty" json:"disableOptions,omitempty"`
	DisableHead    bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty"`

	// Headers 附加的报头信息
	//
	// 一些诸如跨域等报头信息，可以在此作设置。
	//
	// 报头信息可能在其它处理器被修改。
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Static 静态内容，键名为 URL 路径，键值为文件地址
	//
	// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
	// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
	// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
	Static map[string]string `yaml:"static,omitempty" json:"static,omitempty"`

	// AllowedDomains 限定访问域名。
	//
	// 若指定了此值，则只有此列表中指定的域名可以访问当前网页。
	// 诸如 IP 和其它域名的指向将不再启作用。
	//
	// 在 AllowedDomains 中至少存在一个及以上的域名时，Domain
	// 中指定的域名会自动合并到当前列表中。
	// AllowedDomains 为空时，并不会限定域名为 Domain 指定的域名。
	AllowedDomains []string `yaml:"allowedDomains,omitempty" json:"allowedDomains,omitempty"`

	// 应用于 http.Server 的几个变量。
	ReadTimeout       time.Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty"`
	WriteTimeout      time.Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty"`
	IdleTimeout       time.Duration `yaml:"idleTiemout,omitempty" json:"idleTiemout,omitempty"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty"`

	// Compress 表示压缩的相关配置
	//
	// 可以使用 * 作为结尾，同时指定多个，比如：
	// text/* 表示所有以 text/* 开头的 mime-type 类型。
	Compress []string `yaml:"compress,omitempty" json:"compress,omitempty"`

	// 表示关闭整个服务时，需要等待的时间。
	//
	// 若为 0，表示直接关闭，否则会等待该指定的时间，或是在超时时才执行强制关闭。
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout,omitempty" json:"shutdownTimeout,omitempty"`

	// URL 网站的根地址。
	// 一般情况下，如果用到诸如生成 URL 地址什么的，会用到此值。
	//
	// 若为空，则会根据配置文件的内容，生成网站首页地址。
	// 若是 domain 为空，则生成的地址只有路径部分。
	//
	// 用户也可台强制指定一个不同的地址，比如在被反向代理时，
	// 此值可能就和从 Domain、Port 等配置项自动生成的不一样。
	URL     string `yaml:"url,omitempty" json:"url,omitempty"`
	URLPath string `yaml:"-" json:"-"` // URL 的 path 部分
}

// Sanitize 修正可修正的内容，返回不可修正的错误。
func (conf *WebConfig) Sanitize() error {
	if conf.Domain != "" && !is.URL(conf.Domain) && conf.Domain != localhostURL {
		return &config.Error{Field: "domain", Message: "必须是一个 URL"}
	}

	if conf.ReadTimeout < 0 {
		return &config.Error{Field: "readTimeout", Message: "必须大于等于 0"}
	}

	if conf.WriteTimeout < 0 {
		return &config.Error{Field: "writeTimeout", Message: "必须大于等于 0"}
	}

	if conf.IdleTimeout < 0 {
		return &config.Error{Field: "idleTimeout", Message: "必须大于等于 0"}
	}

	if conf.ReadHeaderTimeout < 0 {
		return &config.Error{Field: "readHeaderTimeout", Message: "必须大于等于 0"}
	}

	if conf.MaxHeaderBytes < 0 {
		return &config.Error{Field: "maxHeaderBytes", Message: "必须大于等于 0"}
	}

	if conf.ShutdownTimeout < 0 {
		return &config.Error{Field: "shutdownTimeout", Message: "必须大于等于 0"}
	}

	if err := conf.checkStatic(); err != nil {
		return err
	}

	if err := conf.buildRoot(); err != nil {
		return err
	}
	if err := conf.buildAllowedDomains(); err != nil {
		return err
	}
	if err := conf.buildHTTPS(); err != nil {
		return err
	}

	return conf.buildURL()
}

func (conf *WebConfig) checkStatic() (err error) {
	for url, path := range conf.Static {
		if !isURLPath(url) {
			return &config.Error{
				Field:   "static." + url,
				Message: "必须以 / 开头且不能以 / 结尾",
			}
		}

		if !filepath.IsAbs(path) {
			path, err = filepath.Abs(path)
			if err != nil {
				return &config.Error{Field: "static." + url, Message: err.Error()}
			}
		}

		if !utils.FileExists(path) {
			return &config.Error{Field: "static." + url, Message: "对应的路径不存在"}
		}
		conf.Static[url] = path
	}

	return nil
}

func (conf *WebConfig) buildRoot() error {
	if conf.Root == "/" {
		conf.Root = conf.Root[:0]
	} else if (len(conf.Root) > 0) && !isURLPath(conf.Root) {
		return &config.Error{Field: "root", Message: "必须以 / 开头且不以 / 结尾"}
	}

	return nil
}

func isURLPath(path string) bool {
	return path[0] == '/' && path[len(path)-1] != '/'
}

func (conf *WebConfig) buildHTTPS() error {
	if conf.HTTPS {
		if !utils.FileExists(conf.CertFile) {
			return &config.Error{Field: "certFile", Message: "文件不存在"}
		}

		if !utils.FileExists(conf.KeyFile) {
			return &config.Error{Field: "keyFile", Message: "文件不存在"}
		}
	}

	if conf.Port == 0 {
		if conf.HTTPS {
			conf.Port = 443
		} else {
			conf.Port = 80
		}
	}

	return nil
}

func (conf *WebConfig) buildAllowedDomains() error {
	if len(conf.AllowedDomains) == 0 {
		return nil
	}

	found := false // 确定 domain 是否已经在 allowedDomains 中
	for _, host := range conf.AllowedDomains {
		if !is.URL(host) {
			return &config.Error{
				Field:   "allowedDomains." + host,
				Message: "非法的 URL",
			}
		}

		if host == conf.Domain {
			found = true
		}
	}

	if !found && conf.Domain != "" {
		conf.AllowedDomains = append(conf.AllowedDomains, conf.Domain)
	}

	return nil
}

func (conf *WebConfig) buildURL() error {
	if conf.URL != "" {
		goto PARSE
	}

	// 未指定域名，则只有路径部分
	if conf.Domain == "" {
		conf.URL = conf.Root
		goto PARSE
	}

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

PARSE:
	return conf.parseURL()
}

func (conf *WebConfig) parseURL() error {
	obj, err := url.Parse(conf.URL)
	if err != nil {
		return err
	}

	conf.URLPath = obj.Path

	return nil
}
