// SPDX-License-Identifier: MIT

package logs

import (
	"sync"

	"github.com/issue9/logs/v4"
)

var paramsLogsPool = &sync.Pool{New: func() any { return &ParamsLogs{} }}

type (
	// Logs 日志对象
	Logs struct {
		logs *logs.Logs
	}

	// ParamsLogs 带参数的日志
	ParamsLogs struct {
		logs    *Logs
		ps      map[string]any
		loggers map[Level]Logger
	}
)

func (l *Logs) INFO() Logger { return l.logs.INFO() }

func (l *Logs) WARN() Logger { return l.logs.WARN() }

func (l *Logs) TRACE() Logger { return l.logs.TRACE() }

func (l *Logs) DEBUG() Logger { return l.logs.DEBUG() }

func (l *Logs) ERROR() Logger { return l.logs.ERROR() }

func (l *Logs) FATAL() Logger { return l.logs.FATAL() }

func (l *Logs) NewEntry(lv Level) *Entry { return l.logs.NewEntry(lv) }

// With 构建一个带有指定参数日志对象
func (l *Logs) With(ps map[string]any) *ParamsLogs {
	p := paramsLogsPool.Get().(*ParamsLogs)
	p.logs = l
	p.ps = ps
	p.loggers = make(map[logs.Level]logs.Logger, 6)
	return p
}

func (l *ParamsLogs) INFO() Logger { return l.level(Info) }

func (l *ParamsLogs) TRACE() Logger { return l.level(Trace) }

func (l *ParamsLogs) WARN() Logger { return l.level(Warn) }

func (l *ParamsLogs) DEBUG() Logger { return l.level(Debug) }

func (l *ParamsLogs) ERROR() Logger { return l.level(Error) }

func (l *ParamsLogs) FATAL() Logger { return l.level(Fatal) }

func (l *ParamsLogs) NewEntry(lv Level) *Entry {
	e := l.logs.NewEntry(lv)
	for k, v := range l.ps {
		e.With(k, v)
	}
	return e
}

func (l *ParamsLogs) level(lv Level) Logger {
	if _, found := l.loggers[lv]; !found {
		l.loggers[lv] = l.logs.logs.With(lv, l.ps)
	}
	return l.loggers[lv]
}

// Destroy 回收由 [ParamsLogs] 对象
func Destroy(l *ParamsLogs) { paramsLogsPool.Put(l) }
