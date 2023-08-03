// SPDX-License-Identifier: MIT

// Package logger 错误日志的处理
package logger

import (
	"fmt"
	"go/scanner"

	"github.com/issue9/web"
	"github.com/issue9/web/logs"
	"golang.org/x/mod/modfile"
	"golang.org/x/text/message"
)

type Logger struct {
	logs  logs.Logs
	p     *message.Printer
	count int
}

func New(l logs.Logs, p *message.Printer) *Logger {
	return &Logger{
		logs: l,
		p:    p,
	}
}

// Count 接收到的日志数量
func (l *Logger) Count() int { return l.count }

func (l *Logger) Info(msg web.LocaleStringer) { l.log(logs.Info, msg, "", 0) }

func (l *Logger) Warning(msg web.LocaleStringer) { l.log(logs.Warn, msg, "", 0) }

// 如果 msg 包含了定位信息，则 filename 和 line 将被忽略
func (l *Logger) Error(msg any, filename string, line int) {
	switch e := msg.(type) {
	case web.LocaleStringer:
		l.log(logs.Error, e, filename, line)
	case string:
		l.log(logs.Error, web.Phrase(e), filename, line)
	case fmt.Stringer:
		l.log(logs.Error, web.Phrase(e.String()), filename, line)
	case *scanner.Error:
		l.log(logs.Error, web.Phrase(e.Msg), e.Pos.Filename, e.Pos.Line)
	case scanner.ErrorList:
		for _, err := range e {
			l.Error(err.Msg, err.Pos.Filename, err.Pos.Line)
		}
	case *modfile.Error:
		l.log(logs.Error, web.Phrase(e.Err.Error()), e.Filename, e.Pos.Line)
	case modfile.ErrorList:
		for _, err := range e {
			l.Error(err.Err, err.Filename, err.Pos.Line)
		}
	default:
		l.log(logs.Error, web.Phrase(fmt.Sprint(e)), filename, line)
	}
}

func (l *Logger) log(lv logs.Level, msg web.LocaleStringer, filename string, line int) {
	l.count++

	m := msg.LocaleString(l.p)

	if filename != "" {
		m = web.Phrase("%s at %s:%d", m, filename, line).LocaleString(l.p)
	}

	l.logs.Logger(lv).String(m)
}
