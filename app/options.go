// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"encoding/json"
	"encoding/xml"

	"github.com/issue9/logs/v2"
	"github.com/issue9/mux"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/middleware"
	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/mimetypes"
	"github.com/issue9/web/internal/modules"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
)

var configUnmarshals = map[string]config.UnmarshalFunc{
	".yaml": yaml.Unmarshal,
	".yml":  yaml.Unmarshal,
	".xml":  xml.Unmarshal,
	".json": json.Unmarshal,
}

// Options App 的配置项
type Options struct {
	// 配置文件所在的目录，不能为空。
	// 框架自带的 web.yaml 和 logs.xml 也都将在此目录下。
	Dir string

	// 指定状态下对应的错误处理函数。
	//
	// 若该状态码的处理函数不存在，则会查找键值为 0 的函数，
	// 若依然不存在，则调用 defaultRender
	//
	// 用户也可以通过调用 App.AddErrorHandler 进行添加。
	ErrorHandlers map[int]ErrorHandler

	// 指定使用的中间件。
	//
	// 用户也可以通过 app.AddMiddlewares 进行添加。
	Middlewares []middleware.Middleware

	// 指定 mimetype 的编码函数
	MimetypeMarshals map[string]mimetype.MarshalFunc

	// 指定 mimetype 的解码函数
	MimetypeUnmarshals map[string]mimetype.UnmarshalFunc
}

func (opt *Options) newApp() (*App, error) {
	mgr, err := config.NewManager(opt.Dir)
	if err != nil {
		return nil, err
	}
	for k, v := range configUnmarshals {
		if err := mgr.AddUnmarshal(v, k); err != nil {
			return nil, err
		}
	}

	logs := logs.New()
	if err = logs.InitFromXMLFile(mgr.File(logsFilename)); err != nil {
		return nil, err
	}

	webconf := &webconfig.WebConfig{}
	if err = mgr.LoadFile(configFilename, webconf); err != nil {
		return nil, err
	}

	mt := mimetypes.New()
	if err = mt.AddMarshals(opt.MimetypeMarshals); err != nil {
		return nil, err
	}
	if err = mt.AddUnmarshals(opt.MimetypeUnmarshals); err != nil {
		return nil, err
	}

	middlewares := opt.Middlewares
	if middlewares == nil {
		middlewares = make([]middleware.Middleware, 0, 10)
	}

	errorHandlers := opt.ErrorHandlers
	if errorHandlers == nil {
		errorHandlers = make(map[int]ErrorHandler, 10)
	}

	mux := mux.New(webconf.DisableOptions, false, nil, nil)

	ms, err := modules.New(mux, webconf)
	if err != nil {
		return nil, err
	}

	return &App{
		webConfig:     webconf,
		middlewares:   middlewares,
		mux:           mux,
		closed:        make(chan bool, 1),
		modules:       ms,
		mt:            mt,
		configs:       mgr,
		logs:          logs,
		errorHandlers: errorHandlers,
	}, nil
}
