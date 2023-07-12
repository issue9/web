// SPDX-License-Identifier: MIT

// Package logger 错误日志的处理
package logger

import (
	"go/scanner"

	"golang.org/x/mod/modfile"
)

// 日志类型
const (
	Unknown Type = iota
	Info
	Warning
	Cancelled
	ModSyntax
	GoSyntax  // Go 的语法错误
	DocSyntax // 文档语法错误
	typeSize
)

type Type int8

type Entry struct {
	Filename string
	Line     int
	Msg      any
	Type     Type
}

type Logger struct {
	handler func(*Entry)
	count   int
}

// handler 用于解决如何输出日志对象 [Entry]
func New(handler func(*Entry)) *Logger {
	return &Logger{handler: handler}
}

// Count 接收到的日志数量
func (l *Logger) Count() int { return l.count }

func (l *Logger) Log(t Type, msg any, filename string, line int) {
	l.count++
	l.handler(&Entry{
		Filename: filename,
		Line:     line,
		Msg:      msg,
		Type:     t,
	})
}

func (l *Logger) LogWithoutPos(t Type, msg any) { l.Log(t, msg, "", 0) }

// LogError 将 go 文件解析中的错误输出
//
// filename 和 line 仅在 err 不携带文件信息时才会用到。
func (l *Logger) LogError(t Type, err error, filename string, line int) {
	if se, ok := err.(*scanner.Error); ok {
		l.Log(t, se.Msg, se.Pos.Filename, se.Pos.Line)
	} else if sel, ok := err.(scanner.ErrorList); ok {
		for _, se = range sel {
			l.Log(t, se.Msg, se.Pos.Filename, se.Pos.Line)
		}
	} else if me, ok := err.(*modfile.Error); ok {
		l.Log(t, me.Err, filename, me.Pos.Line)
	} else if mel, ok := err.(modfile.ErrorList); ok {
		for _, e := range mel {
			l.Log(t, e.Err, filename, e.Pos.Line)
		}
	} else {
		l.Log(t, se, filename, line)
	}
}
