// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package config 提供了程序对自身的配置文件的操作能力。
//
// NOTE: 所有需要写入到配置文件的配置项，都应该在此定义。
package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/issue9/utils"
)

// 端口的定义
const (
	HTTPPort  = ":80"
	HTTPSPort = ":443"
)

// 当启用 HTTPS 时，对 80 端口的处理方式。
const (
	HTTPStateDisabled = "disable"  // 禁止监听 80 端口
	HTTPStateListen   = "listen"   // 监听 80 端口，与 HTTPS 相同的方式处理
	HTTPStateRedirect = "redirect" // 监听 80 端口，并重定向到 HTTPS
)

// Config 系统配置文件。
type Config struct {
	// 基本
	HTTPS       bool              `json:"https,omitempty"`     // 是否启用 HTTPS
	HTTPState   string            `json:"httpState,omitempty"` // 80 端口的状态，仅在 HTTPS 为 true 时启作用
	CertFile    string            `json:"certFile,omitempty"`  // 当 https 为 true 时，此值为必填
	KeyFile     string            `json:"keyFile,omitempty"`   // 当 https 为 true 时，此值为必填
	Port        string            `json:"port,omitempty"`      // 端口，不指定，默认为 80 或是 443
	Headers     map[string]string `json:"headers,omitempty"`   // 附加的头信息，头信息可能在其它地方被修改
	Static      map[string]string `json:"static,omitempty"`    // 静态内容，键名为 URL 路径，键值为文件地址
	ContentType string            `json:"contentType"`         // 默认的编码类型

	// 性能
	ReadTimeout  time.Duration `json:"readTimeout,omitempty"`  // http.Server.ReadTimeout 的值，单位：秒
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"` // http.Server.WriteTimeout 的值，单位：秒
	Pprof        string        `json:"pprof,omitempty"`        // 指定 pprof 地址

	// Envelope
	Envelope *Envelope `json:"envelope,omitempty"`
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

	if len(conf.Port) == 0 {
		if conf.HTTPS {
			conf.Port = HTTPSPort
		} else {
			conf.Port = HTTPPort
		}
	}
	if conf.Port[0] != ':' {
		conf.Port = ":" + conf.Port
	}

	if len(conf.HTTPState) > 0 && (conf.HTTPState != HTTPStateDisabled &&
		conf.HTTPState != HTTPStateListen &&
		conf.HTTPState != HTTPStateRedirect) {
		return nil, errors.New("无效的 httpState 值")
	}

	if conf.HTTPS {
		if !utils.FileExists(conf.CertFile) {
			return nil, errors.New("certFile 所指的文件并不存在")
		}
		if !utils.FileExists(conf.KeyFile) {
			return nil, errors.New("keyFile 所指的文件并不存在")
		}
	}

	if len(conf.ContentType) == 0 {
		return nil, errors.New("contentType 未指定")
	}

	return conf, nil
}
