// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/issue9/is"
	"github.com/issue9/utils"
)

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

// Config 系统配置文件。
type Config struct {
	// 基本
	HTTPS     bool              `json:"https"`     // 是否启用 HTTPS
	HTTPState string            `json:"httpState"` // 80 端口的状态，仅在 HTTPS 为 true 时启作用
	CertFile  string            `json:"certFile"`  // 当 https 为 true 时，此值为必填
	KeyFile   string            `json:"keyFile"`   // 当 https 为 true 时，此值为必填
	Port      string            `json:"port"`      // 端口，不指定，默认为 80 或是 443
	Headers   map[string]string `json:"headers"`   // 附加的头信息，头信息可能在其它地方被修改
	Static    map[string]string `json:"static"`    // 静态内容，键名为 URL 路径，键值为文件地址
	Options   bool              `json:"options"`   // 是否启用 OPTIONS 请求
	Version   string            `json:"version"`   // 限定版本
	Hosts     []string          `json:"hosts"`     // 限定访问域名。仅需指定域名，端口及其它任何信息不需要指定

	// 性能
	ReadTimeout  time.Duration `json:"readTimeout"`  // http.Server.ReadTimeout 的值，单位：纳秒
	WriteTimeout time.Duration `json:"writeTimeout"` // http.Server.WriteTimeout 的值，单位：纳秒
	Pprof        string        `json:"pprof"`        // 指定 pprof 地址
}

// DefaultConfig 返回一个默认的 Config
func DefaultConfig() *Config {
	return &Config{
		HTTPS:     false,
		HTTPState: httpStateDisabled,
		CertFile:  "",
		KeyFile:   "",
		Port:      ":80",
		Headers:   nil,
		Static:    nil,
		Options:   true,
		Version:   "",
		Hosts:     []string{},

		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Pprof:        "/debub/pprof",
	}
}

// Sanitize 检测各个值是否正常
func (conf *Config) Sanitize() error {
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
		return errors.New("port 必须指定")
	} else if conf.Port[0] != ':' {
		conf.Port = ":" + conf.Port
	}

	if len(conf.Hosts) > 0 {
		for _, host := range conf.Hosts {
			if !is.URL(host) {
				return fmt.Errorf("conf.Hosts 中的 %v 为非法的 URL", host)
			}
		}
	}

	if len(conf.Pprof) > 0 && conf.Pprof[len(conf.Pprof)-1] != '/' {
		conf.Pprof = conf.Pprof + "/"
	}

	if conf.ReadTimeout < 0 {
		return errors.New("readTimeout 必须大于等于 0")
	}

	if conf.WriteTimeout < 0 {
		return errors.New("writeTimeout 必须大于等于 0")
	}
	return nil
}
