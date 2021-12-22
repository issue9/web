// SPDX-License-Identifier: MIT

// Package web 一个微型的 web 框架
package web

import (
	"io/fs"
	"net/http"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/config"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.44.0"

type (
	Server       = server.Server
	Context      = server.Context
	Options      = server.Options
	Filter       = server.Filter
	HandlerFunc  = server.HandlerFunc
	Responser    = server.Responser
	Router       = server.Router
	ResultFields = server.ResultFields
	Result       = server.Result
	Files        = serialization.Files

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer

	// OptionsFunc 用于对 Options 对象进行修改
	OptionsFunc func(*Options)
)

// LoadServer 从配置文件初始化 Server 对象
//
// files 指定了用于加载本地化的方法，同时也用于加载配置文件；
// conf 用于指定项目的配置文件，根据扩展由 serialization.Files 负责在 f 查找文件加载；
// o 用于在初始化 Server 之前，加载配置文件之后，对 *Options 进行一次修改；
func LoadServer(name, version string, files *Files, f fs.FS, conf string, o OptionsFunc) (*Server, error) {
	opt, err := config.NewOptions(files, f, conf)
	if err != nil {
		return nil, err
	}
	if o != nil {
		o(opt)
	}
	return NewServer(name, version, opt)
}

// NewServer 从 Options 初始化 Server 对象
func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

func Status(status int) Responser { return server.Status(status) }

func Object(status int, body interface{}, headers map[string]string) Responser {
	return server.Object(status, body, headers)
}

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...interface{}) LocaleStringer {
	return localeutil.Phrase(key, v...)
}

func Created(v interface{}, location string) Responser {
	if location == "" {
		return Object(http.StatusCreated, v, nil)
	}

	return Object(http.StatusCreated, v, map[string]string{
		"Location": location,
	})
}

// OK 返回 200 状态码下的对象
func OK(v interface{}) Responser { return Object(http.StatusOK, v, nil) }

func NotFound() Responser { return Status(http.StatusNotFound) }

func NoContent() Responser { return Status(http.StatusNoContent) }

func NotImplemented() Responser { return Status(http.StatusNotImplemented) }
