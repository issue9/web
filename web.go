// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个模块化的微形 web 框架。
//  m, err := web.NewModule("m1")
//  m.Get("/", ...).
//    Post("/", ...)
//
//
//  conf := &web.Config{
//      Pprof:  "/debug/pprof",
//      Before: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){...}),
//      After:  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){...}),
//  }
//  web.Run(conf) // 开始监听端口
//
// NOTE: web 依赖 github.com/issue9/logs 包作日志输出，请确保已经正确初始化该包。
package web

import (
	"net/http"

	"github.com/issue9/logs"
)

// Run 运行路由，执行监听程序。
func Run(conf *Config) error {
	return conf.run()
}

func message(r *http.Request, v []interface{}) []interface{} {
	if r == nil {
		return v
	}

	return append(v, "@", r.URL)
}

// Critical 相当于调用了 logs.Critical，外加一些调用者的详细信息
func Critical(r *http.Request, v ...interface{}) {
	logs.Critical(message(r, v)...)
}

// Criticalf 相当于调用了 logs.Criticalf，外加一些调用者的详细信息
func Criticalf(r *http.Request, format string, v ...interface{}) {
	logs.Criticalf(format, message(r, v)...)
}

// Error 相当于调用了 logs.Error，外加一些调用者的详细信息
func Error(r *http.Request, v ...interface{}) {
	logs.Error(message(r, v)...)
}

// Errorf 相当于调用了 logs.Errorf，外加一些调用者的详细信息
func Errorf(r *http.Request, format string, v ...interface{}) {
	logs.Errorf(format, message(r, v)...)
}

// Debug 相当于调用了 logs.Debug，外加一些调用者的详细信息
func Debug(r *http.Request, v ...interface{}) {
	logs.Debug(message(r, v)...)
}

// Debugf 相当于调用了 logs.Debugf，外加一些调用者的详细信息
func Debugf(r *http.Request, format string, v ...interface{}) {
	logs.Debugf(format, message(r, v)...)
}

// Trace 相当于调用了 logs.Trace，外加一些调用者的详细信息
func Trace(r *http.Request, v ...interface{}) {
	logs.Trace(message(r, v)...)
}

// Tracef 相当于调用了 logs.Tracef，外加一些调用者的详细信息
func Tracef(r *http.Request, format string, v ...interface{}) {
	logs.Tracef(format, message(r, v)...)
}

// Warn 相当于调用了 logs.Warn，外加一些调用者的详细信息
func Warn(r *http.Request, v ...interface{}) {
	logs.Warn(message(r, v)...)
}

// Warnf 相当于调用了 logs.Warnf，外加一些调用者的详细信息
func Warnf(r *http.Request, format string, v ...interface{}) {
	logs.Warnf(format, message(r, v)...)
}

// Info 相当于调用了 logs.Info，外加一些调用者的详细信息
func Info(r *http.Request, v ...interface{}) {
	logs.Info(message(r, v)...)
}

// Infof 相当于调用了 logs.Infof，外加一些调用者的详细信息
func Infof(r *http.Request, format string, v ...interface{}) {
	logs.Infof(format, message(r, v)...)
}
