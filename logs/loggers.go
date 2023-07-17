// SPDX-License-Identifier: MIT

package logs

import (
	"log"
	"sync"

	"github.com/issue9/logs/v5"
	"github.com/issue9/logs/v5/writers"
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

// New 声明日志实例
func New(opt *Options) (*Logs, error) {
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

	if l.IsEnable(opt.StdLevel) {
		sl := l.Logger(opt.StdLevel)
		log.SetOutput(writers.WriteFunc(func(data []byte) (int, error) {
			sl.String(string(data))
			return len(data), nil
		}))
	}

	return &Logs{logs: l}, nil
}

func (l *Logs) INFO() Logger { return l.logs.INFO() }

func (l *Logs) WARN() Logger { return l.logs.WARN() }

func (l *Logs) TRACE() Logger { return l.logs.TRACE() }

func (l *Logs) DEBUG() Logger { return l.logs.DEBUG() }

func (l *Logs) ERROR() Logger { return l.logs.ERROR() }

func (l *Logs) FATAL() Logger { return l.logs.FATAL() }

func (l *Logs) Logger(lv Level) Logger { return l.logs.Logger(lv) }

func (l *Logs) NewRecord(lv Level) *Record { return l.logs.NewRecord(lv) }

// With 构建一个带有指定参数日志对象
func (l *Logs) With(ps map[string]any) *ParamsLogs {
	p := paramsLogsPool.Get().(*ParamsLogs)
	p.logs = l
	p.ps = ps
	p.loggers = make(map[Level]Logger, 6)
	return p
}

func (l *ParamsLogs) INFO() Logger { return l.Logger(Info) }

func (l *ParamsLogs) TRACE() Logger { return l.Logger(Trace) }

func (l *ParamsLogs) WARN() Logger { return l.Logger(Warn) }

func (l *ParamsLogs) DEBUG() Logger { return l.Logger(Debug) }

func (l *ParamsLogs) ERROR() Logger { return l.Logger(Error) }

func (l *ParamsLogs) FATAL() Logger { return l.Logger(Fatal) }

func (l *ParamsLogs) NewRecord(lv Level) *Record {
	e := l.logs.NewRecord(lv)
	for k, v := range l.ps {
		e.With(k, v)
	}
	return e
}

func (l *ParamsLogs) Logger(lv Level) Logger {
	if _, found := l.loggers[lv]; !found {
		l.loggers[lv] = l.logs.logs.With(lv, l.ps)
	}
	return l.loggers[lv]
}

// DestroyParamsLogs 回收由 [ParamsLogs] 对象
//
// 这是一个非必须的方法，调用可能会有一定的性能提升。
func DestroyParamsLogs(l *ParamsLogs) { paramsLogsPool.Put(l) }
