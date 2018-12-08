// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/issue9/logs/v2"
	"github.com/issue9/mux"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/middleware"
	"github.com/issue9/web/config"
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

	// 指定使用的中间件。
	//
	// 用户也可以通过 app.AddMiddlewares 进行添加。
	Middlewares []middleware.Middleware
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

	middlewares := opt.Middlewares
	if middlewares == nil {
		middlewares = make([]middleware.Middleware, 0, 10)
	}

	mux := mux.New(webconf.DisableOptions, false, notFound, methodNotAllowed)

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
		mt:            mimetype.New(),
		configs:       mgr,
		logs:          logs,
		errorHandlers: make(map[int]ErrorHandler, 10),
	}, nil
}

func notFound(w http.ResponseWriter, r *http.Request) {
	ExitContext(http.StatusNotFound)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ExitContext(http.StatusMethodNotAllowed)
}
