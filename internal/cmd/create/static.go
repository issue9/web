// SPDX-License-Identifier: MIT

package create

import (
	"github.com/issue9/logs/v2/config"

	"github.com/issue9/web"
)

// go.mod
const gomod = `module %s

require github.com/issue9/web v%s`

// main.go
const maingo = `// 内容由 web 自动生成，可根据需求自由修改！

package main

import (
    "encoding/json"
    "encoding/xml"

    "github.com/issue9/web"
)

const appconfig = "./appconfig"

func main() {
    web.Classic(appconfig)

    // 所有的模块初始化在此函数
    initModules()

    web.Fatal(2, web.Serve())
}

// initModules 所有模块的初始化操作可在此处进行。
func initModules() {
	// TODO
}
`

var webconf = &web.Config{
	Root: "http://localhost:8080/",
}

var logsconf = &config.Config{
	Items: map[string]*config.Config{
		"info": {
			Attrs: map[string]string{
				"prefix": "INFO",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"buffer": {
					Attrs: map[string]string{"size": "100"},
					Items: map[string]*config.Config{
						"rotate": {
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

		"debug": {
			Attrs: map[string]string{
				"prefix": "DEBUG",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": {
					Attrs: map[string]string{
						"filename": "debug-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"warn": {
			Attrs: map[string]string{
				"prefix": "WARN",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": {
					Attrs: map[string]string{
						"filename": "warn-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"trace": {
			Attrs: map[string]string{
				"prefix": "TRACE",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": {
					Attrs: map[string]string{
						"filename": "trace-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"error": {
			Attrs: map[string]string{
				"prefix": "ERROR",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": {
					Attrs: map[string]string{
						"filename": "error-%Y%m%d.%i.log",
						"dir":      "./logs",
						"size":     "5m",
					},
				},
			},
		},

		"critical": {
			Attrs: map[string]string{
				"prefix": "CRITICAL",
				"flag":   "log.Llongfile|log.Ldate|log.Ltime",
			},
			Items: map[string]*config.Config{
				"rotate": {
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
