// SPDX-License-Identifier: MIT

// Package logs 日志操作
package logs

import (
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
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
	Level    = logs.Level
	Handler  = logs.Handler
	Record   = logs.Record
	Recorder = logs.Recorder
	Logger   = logs.Logger
	Buffer   = logs.Buffer
	Attr     = logs.Attr

	// Logs 日志系统接口
	Logs interface {
		INFO() *Logger

		WARN() *Logger

		TRACE() *Logger

		DEBUG() *Logger

		ERROR() *Logger

		FATAL() *Logger

		Logger(Level) *Logger

		NewRecord() *Record

		// New 构建一个带有指定属性的 [Logs]
		New(map[string]any) Logs

		// AppendAttrs 添加共同属性
		AppendAttrs(map[string]any)

		// Free 回收当前对象
		//
		// 大部分时候不需要主动调用此方法。但是一旦调用，
		// 表示该对象已经不可用。
		Free()
	}

	defaultLogs struct {
		logs *logs.Logs
	}

	withLogs struct {
		freed   bool
		logs    *defaultLogs
		attrs   map[string]any
		loggers map[Level]*Logger
	}
)

// New 声明日志实例
//
// p 关联的本地对象，[Logger.Error] 和 [Logger.Print] 等的输出受此影响，可以为空，具体可参考 [logs.WithLocale]。
func New(p *localeutil.Printer, o *Options) (Logs, error) {
	o, err := optionsSanitize(o)
	if err != nil {
		return nil, err
	}

	oo := make([]logs.Option, 0, 5)

	oo = append(oo, logs.WithLocale(p))

	if o.Location {
		oo = append(oo, logs.WithLocation(true))
	}
	if o.Created != "" {
		oo = append(oo, logs.WithCreated(o.Created))
	}
	if o.StackError {
		oo = append(oo, logs.WithDetail(true))
	}

	if o.Std {
		oo = append(oo, logs.WithStd())
	}

	if len(o.Levels) > 0 {
		oo = append(oo, logs.WithLevels(o.Levels...))
	}

	return &defaultLogs{logs: logs.New(o.Handler, oo...)}, nil
}

func (l *defaultLogs) INFO() *Logger { return l.logs.INFO() }

func (l *defaultLogs) WARN() *Logger { return l.logs.WARN() }

func (l *defaultLogs) TRACE() *Logger { return l.logs.TRACE() }

func (l *defaultLogs) DEBUG() *Logger { return l.logs.DEBUG() }

func (l *defaultLogs) ERROR() *Logger { return l.logs.ERROR() }

func (l *defaultLogs) FATAL() *Logger { return l.logs.FATAL() }

func (l *defaultLogs) AppendAttrs(attrs map[string]any) { l.logs.AppendAttrs(attrs) }

func (l *defaultLogs) Logger(lv Level) *Logger { return l.logs.Logger(lv) }

func (l *defaultLogs) NewRecord() *Record { return l.logs.NewRecord() }

// With 构建一个带有指定参数日志对象
func (l *defaultLogs) New(attrs map[string]any) Logs {
	p := withLogsPool.Get().(*withLogs)
	p.freed = false
	p.logs = l
	p.attrs = attrs
	p.loggers = make(map[Level]*Logger, 6)
	return p
}

func (l *defaultLogs) Free() {}

func (l *withLogs) INFO() *Logger { return l.Logger(Info) }

func (l *withLogs) TRACE() *Logger { return l.Logger(Trace) }

func (l *withLogs) WARN() *Logger { return l.Logger(Warn) }

func (l *withLogs) DEBUG() *Logger { return l.Logger(Debug) }

func (l *withLogs) ERROR() *Logger { return l.Logger(Error) }

func (l *withLogs) FATAL() *Logger { return l.Logger(Fatal) }

func (l *withLogs) AppendAttrs(attrs map[string]any) {
	for _, ll := range l.loggers {
		if ll != nil {
			ll.AppendAttrs(attrs)
		}
	}
	for k, v := range attrs {
		l.attrs[k] = v
	}
}

func (l *withLogs) NewRecord() *Record { return l.logs.NewRecord() }

func (l *withLogs) Logger(lv Level) *Logger {
	if _, found := l.loggers[lv]; !found {
		l.loggers[lv] = l.logs.Logger(lv).New(l.attrs)
	}
	return l.loggers[lv]
}

func (l *withLogs) New(attrs map[string]any) Logs {
	for k, v := range l.attrs {
		attrs[k] = v
	}
	return l.logs.New(attrs)
}

func (l *withLogs) Free() {
	if l.freed {
		return
	}
	l.freed = true
	withLogsPool.Put(l)
}

func NewBuffer(detail bool) *Buffer { return logs.NewBuffer(detail) }
