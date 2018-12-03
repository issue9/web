// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"github.com/issue9/logs/v2"
	"github.com/issue9/mux"

	"github.com/issue9/middleware"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/modules"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
)

// Config 配置项
type Config struct {
	Dir                string
	ErrorHandlers      map[int]ErrorHandler
	Compresses         map[string]compress.WriterFunc
	Middlewares        []middleware.Middleware
	ConfigUnmarshals   map[string]config.UnmarshalFunc
	MimetypeMarshals   map[string]mimetype.MarshalFunc
	MimetypeUnmarshals map[string]mimetype.UnmarshalFunc
}

// New 声明一个新的 App 实例
//
// 日志系统会在此处初始化。
func (conf *Config) newApp() (*App, error) {
	mgr, err := config.NewManager(conf.Dir)
	if err != nil {
		return nil, err
	}
	for k, v := range conf.ConfigUnmarshals {
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

	mt := mimetype.New()
	if err = mt.AddMarshals(conf.MimetypeMarshals); err != nil {
		return nil, err
	}
	if err = mt.AddUnmarshals(conf.MimetypeUnmarshals); err != nil {
		return nil, err
	}

	middlewares := conf.Middlewares
	if middlewares == nil {
		middlewares = make([]middleware.Middleware, 0, 10)
	}

	errorHandlers := conf.ErrorHandlers
	if errorHandlers == nil {
		errorHandlers = make(map[int]ErrorHandler, 10)
	}

	compresses := conf.Compresses
	if compresses == nil {
		compresses = make(map[string]compress.WriterFunc, 10)
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
		compresses:    compresses,
	}, nil
}
