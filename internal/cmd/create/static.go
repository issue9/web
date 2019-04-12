// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package create

import (
	"github.com/issue9/logs/v2/config"

	"github.com/issue9/web/internal/webconfig"
)

// go.mod
const gomod = `module %s

require github.com/issue9/web v%s`

// main.go
const maingo = `// 内容由 web 自动生成，可根据需求自由修改！

package main

const appconfig = "./appconfig"

import (
    "encoding/json"
    "encoding/xml"

    "github.com/issue9/web"

    "%s"
)

func main() {
    web.Classic(appconfig)

    // 所有的模块初始化在此函数
    modules.Init()

    web.Fatal(2, web.Serve())
}
`

const modulesgo = `// 内容由 web 自动生成，可根据需求自由修改！

// Package modules 完成所有模块的初始化
package modules

// Init 所有模块的初始化操作可在此处进行。
func Init() {
    // TODO
}
`

var webconf = &webconfig.WebConfig{
	HTTPS:  false,
	Domain: "localhost",
	Port:   8080,
	Logs:   "logs.yaml",
}

var logsconf = &config.Config{
	Items: map[string]*config.Config{
		"info": &config.Config{
			Attrs: map[string]string{
				"prefix": "INFO",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"buffer": &config.Config{
					Attrs: map[string]string{"size": "100"},
					Items: map[string]*config.Config{
						"rotate": &config.Config{
							Attrs: map[string]string{
								"filename": "info-%Y%m%d.%i.log",
								"dir":      "./logs",
								"size":     "5m",
							},
						},
					},
				},
			},
		},

		"debug": &config.Config{
			Attrs: map[string]string{
				"prefix": "DEBUG",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": &config.Config{
					Attrs: map[string]string{
						"filename": "debug-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"warn": &config.Config{
			Attrs: map[string]string{
				"prefix": "WARN",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": &config.Config{
					Attrs: map[string]string{
						"filename": "warn-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"trace": &config.Config{
			Attrs: map[string]string{
				"prefix": "TRACE",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": &config.Config{
					Attrs: map[string]string{
						"filename": "trace-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"error": &config.Config{
			Attrs: map[string]string{
				"prefix": "ERROR",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": &config.Config{
					Attrs: map[string]string{
						"filename": "error-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"critical": &config.Config{
			Attrs: map[string]string{
				"prefix": "CRITICAL",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": &config.Config{
					Attrs: map[string]string{
						"filename": "critical-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},
	},
}
