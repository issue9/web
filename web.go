// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// web 一个模块化的微形 web 框架。
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
	"fmt"
	"net/http"
	"strings"

	"github.com/issue9/logs"
)

// Run 运行路由，执行监听程序。
func Run(conf *Config) error {
	return conf.run()
}

func message(r *http.Request, v []interface{}) string {
	if r != nil {
		v = append(v, "@", r.URL)
	}

	return fmt.Sprintln(v...)
}

func messagef(r *http.Request, format string, v []interface{}) string {
	if r != nil {
		format = format + "@" + r.URL.String()
	}

	return fmt.Sprintf(format, v...)
}

// Critical 相当于调用了 logs.Critical，外加一些调用者的详细信息
func Critical(r *http.Request, v ...interface{}) {
	logs.CRITICAL().Output(2, message(r, v))
}

// Criticalf 相当于调用了 logs.Criticalf，外加一些调用者的详细信息
func Criticalf(r *http.Request, format string, v ...interface{}) {
	logs.CRITICAL().Output(2, messagef(r, format, v))
}

// Error 相当于调用了 logs.Error，外加一些调用者的详细信息
func Error(r *http.Request, v ...interface{}) {
	logs.ERROR().Output(2, message(r, v))
}

// Errorf 相当于调用了 logs.Errorf，外加一些调用者的详细信息
func Errorf(r *http.Request, format string, v ...interface{}) {
	logs.ERROR().Output(2, messagef(r, format, v))
}

// Debug 相当于调用了 logs.Debug，外加一些调用者的详细信息
func Debug(r *http.Request, v ...interface{}) {
	logs.DEBUG().Output(2, message(r, v))
}

// Debugf 相当于调用了 logs.Debugf，外加一些调用者的详细信息
func Debugf(r *http.Request, format string, v ...interface{}) {
	logs.DEBUG().Output(2, messagef(r, format, v))
}

// Trace 相当于调用了 logs.Trace，外加一些调用者的详细信息
func Trace(r *http.Request, v ...interface{}) {
	logs.TRACE().Output(2, message(r, v))
}

// Tracef 相当于调用了 logs.Tracef，外加一些调用者的详细信息
func Tracef(r *http.Request, format string, v ...interface{}) {
	logs.TRACE().Output(2, messagef(r, format, v))
}

// Warn 相当于调用了 logs.Warn，外加一些调用者的详细信息
func Warn(r *http.Request, v ...interface{}) {
	logs.WARN().Output(2, message(r, v))
}

// Warnf 相当于调用了 logs.Warnf，外加一些调用者的详细信息
func Warnf(r *http.Request, format string, v ...interface{}) {
	logs.WARN().Output(2, messagef(r, format, v))
}

// Info 相当于调用了 logs.Info，外加一些调用者的详细信息
func Info(r *http.Request, v ...interface{}) {
	logs.INFO().Output(2, message(r, v))
}

// Infof 相当于调用了 logs.Infof，外加一些调用者的详细信息
func Infof(r *http.Request, format string, v ...interface{}) {
	logs.INFO().Output(2, messagef(r, format, v))
}

// ResultFields 从报头中获取 X-Result-Fields 的相关内容。
//
// allow 表示所有允许出现的字段名称。
// 当请求头中未包含 X-Result-Fields 时，返回 nil, true；
// 当请求头包含不允许(不包含在 allow 参数中)的字段是，返回该这些字段，第三个返回参数被设置为 false；
// 否则返回 X-Reslt-Fields 的指定的所有字段，第二个参数返回 true；
// 其它情况返回 nil, false。
func ResultFields(r *http.Request, allow []string) ([]string, bool) {
	if r.Method != http.MethodGet {
		return nil, false
	}

	fields := r.Header.Get("X-Result-Fields")
	if len(fields) == 0 {
		return nil, true
	}

	isAllow := func(field string) bool {
		for _, f1 := range allow {
			if f1 == field {
				return true
			}
		}
		return false
	}

	fs := strings.Split(fields, ",")
	errFields := make([]string, 0, len(fs))

	for index, field := range fs {
		field = strings.TrimSpace(field)
		fs[index] = field

		if isAllow(field) {
			continue
		}
		errFields = append(errFields, field)
	}

	if len(errFields) > 0 {
		return errFields, false
	}

	return fs, true
}
