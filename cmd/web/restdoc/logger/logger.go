// SPDX-License-Identifier: MIT

// Package logger 错误日志的处理
package logger

import (
	"fmt"
	"go/scanner"

	"github.com/issue9/logs/v7"
	"github.com/issue9/web"
	"golang.org/x/mod/modfile"
)

type Logger struct {
	logs  *logs.Logs
	count int
}

func New(l *logs.Logs) *Logger { return &Logger{logs: l} }

// Count 接收到的日志数量
func (l *Logger) Count() int { return l.count }

// Info 输出提示信息
func (l *Logger) Info(msg any) { l.log(logs.LevelInfo, msg, "", 0) }

// Warning 输出警告信息
func (l *Logger) Warning(msg any) { l.log(logs.LevelWarn, msg, "", 0) }

// Error 输出错误信息
//
// 如果 msg 包含了定位信息，则 filename 和 line 将被忽略
func (l *Logger) Error(msg any, filename string, line int) {
	l.log(logs.LevelError, msg, filename, line)
}

// Fatal 输出致命错误
func (l *Logger) Fatal(msg any) {
	l.log(logs.LevelFatal, msg, "", 0)
}

func (l *Logger) log(lv logs.Level, msg any, filename string, line int) {
	var m web.LocaleStringer
	switch e := msg.(type) {
	case web.LocaleStringer:
		m = e
	case string:
		m = web.Phrase(e)
	case fmt.Stringer:
		m = web.Phrase(e.String())
	case *scanner.Error:
		filename = e.Pos.Filename
		line = e.Pos.Line
		m = web.Phrase(e.Msg)
	case scanner.ErrorList:
		for _, err := range e {
			l.log(lv, err.Msg, err.Pos.Filename, err.Pos.Line)
		}
		return
	case *modfile.Error:
		m = web.Phrase(e.Err.Error())
		filename = e.Filename
		line = e.Pos.Line
	case modfile.ErrorList:
		for _, err := range e {
			l.log(lv, err.Err, err.Filename, err.Pos.Line)
		}
		return
	default:
		m = web.Phrase(fmt.Sprint(e))
	}

	l.count++ // 只有真正输出时，才需要+1。

	if filename != "" {
		m = web.Phrase("%s at %s:%d", m, filename, line)
	}

	l.logs.Logger(lv).LocaleString(m)
}
