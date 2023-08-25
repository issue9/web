// SPDX-License-Identifier: MIT

// Package logs 日志操作
package logs

import (
	"sync"

	"github.com/issue9/logs/v5"
)

// 日志的类别
const (
	Info  = logs.LevelInfo
	Trace = logs.LevelTrace
	Warn  = logs.LevelWarn
	Debug = logs.LevelDebug
	Error = logs.LevelError
	Fatal = logs.LevelFatal
)

var withLogsPool = &sync.Pool{New: func() any { return &withLogs{} }}

type (
	Level      = logs.Level
	Handler    = logs.Handler
	HandleFunc = logs.HandleFunc
	Record     = logs.Record
	Logger     = logs.Logger

	// Logs 日志系统接口
	Logs interface {
		INFO() Logger

		WARN() Logger

		TRACE() Logger

		DEBUG() Logger

		ERROR() Logger

		FATAL() Logger

		Logger(Level) Logger

		NewRecord(Level) *Record

		// With 构建一个带有指定参数的日志对象
		With(ps map[string]any) Logs
	}

	defaultLogs struct {
		logs *logs.Logs
	}

	withLogs struct {
		logs    *defaultLogs
		ps      map[string]any
		loggers map[Level]Logger
	}
)

// New 声明日志实例
func New(opt *Options) (Logs, error) {
	opt, err := optionsSanitize(opt)
	if err != nil {
		return nil, err
	}

	o := make([]logs.Option, 0, 3)
	if opt.Caller {
		o = append(o, logs.Caller)
	}
	if opt.Created {
		o = append(o, logs.Created)
	}

	if opt.Handler == nil {
		opt.Handler = NewNopHandler()
	}

	l := logs.New(opt.Handler, o...)
	l.Enable(opt.Levels...)

	if opt.Std {
		setStdDefault(l)
	}

	return &defaultLogs{logs: l}, nil
}

func (l *defaultLogs) INFO() Logger { return l.logs.INFO() }

func (l *defaultLogs) WARN() Logger { return l.logs.WARN() }

func (l *defaultLogs) TRACE() Logger { return l.logs.TRACE() }

func (l *defaultLogs) DEBUG() Logger { return l.logs.DEBUG() }

func (l *defaultLogs) ERROR() Logger { return l.logs.ERROR() }

func (l *defaultLogs) FATAL() Logger { return l.logs.FATAL() }

func (l *defaultLogs) Logger(lv Level) Logger { return l.logs.Logger(lv) }

func (l *defaultLogs) NewRecord(lv Level) *Record { return l.logs.NewRecord(lv) }

// With 构建一个带有指定参数日志对象
func (l *defaultLogs) With(ps map[string]any) Logs {
	p := withLogsPool.Get().(*withLogs)
	p.logs = l
	p.ps = ps
	p.loggers = make(map[Level]Logger, 6)
	return p
}

func (l *withLogs) INFO() Logger { return l.Logger(Info) }

func (l *withLogs) TRACE() Logger { return l.Logger(Trace) }

func (l *withLogs) WARN() Logger { return l.Logger(Warn) }

func (l *withLogs) DEBUG() Logger { return l.Logger(Debug) }

func (l *withLogs) ERROR() Logger { return l.Logger(Error) }

func (l *withLogs) FATAL() Logger { return l.Logger(Fatal) }

func (l *withLogs) NewRecord(lv Level) *Record {
	e := l.logs.NewRecord(lv)
	for k, v := range l.ps {
		e.With(k, v)
	}
	return e
}

func (l *withLogs) Logger(lv Level) Logger {
	if _, found := l.loggers[lv]; !found {
		l.loggers[lv] = l.logs.logs.With(lv, l.ps)
	}
	return l.loggers[lv]
}

func (l *withLogs) With(ps map[string]any) Logs {
	for k, v := range l.ps {
		ps[k] = v
	}
	return l.logs.With(ps)
}

// DestroyWithLogs 回收 [Logs.With] 创建的对象
//
// 这是一个非必须的方法，调用可能会有一定的性能提升。需要确保 l 类型的正确性！
func DestroyWithLogs(l Logs) { withLogsPool.Put(l) }
