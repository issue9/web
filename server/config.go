// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import "time"

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
	HTTPS     bool              `json:"https,omitempty"`     // 是否启用 HTTPS
	HTTPState string            `json:"httpState,omitempty"` // 80 端口的状态，仅在 HTTPS 为 true 时启作用
	CertFile  string            `json:"certFile,omitempty"`  // 当 https 为 true 时，此值为必填
	KeyFile   string            `json:"keyFile,omitempty"`   // 当 https 为 true 时，此值为必填
	Port      string            `json:"port,omitempty"`      // 端口，不指定，默认为 80 或是 443
	Headers   map[string]string `json:"headers,omitempty"`   // 附加的头信息，头信息可能在其它地方被修改
	Static    map[string]string `json:"static,omitempty"`    // 静态内容，键名为 URL 路径，键值为文件地址
	Options   bool              `json:"options,omitempty"`   // 是否启用 OPTIONS 请求

	// 性能
	ReadTimeout  time.Duration `json:"readTimeout,omitempty"`  // http.Server.ReadTimeout 的值，单位：秒
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"` // http.Server.WriteTimeout 的值，单位：秒
	Pprof        string        `json:"pprof,omitempty"`        // 指定 pprof 地址
}
