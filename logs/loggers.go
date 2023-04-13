// SPDX-License-Identifier: MIT

package logs

import (
	"log"
	"sync"

	"github.com/issue9/logs/v4"
	"golang.org/x/text/message"
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

	loggerWriter struct {
		l logs.Logger
	}
)

func (w *loggerWriter) Write(data []byte) (int, error) {
	w.l.String(string(data))
	return len(data), nil
}

// New 声明日志实例
func New(opt *Options, p *message.Printer) (*Logs, error) {
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

	if p != nil {
		o = append(o, logs.Print(newPrinter(p)))
	}

	if opt.Writer == nil {
		opt.Writer = NewNopWriter()
	}

	l := logs.New(opt.Writer, o...)
	l.Enable(opt.Levels...)

	if l.IsEnable(opt.StdLevel) {
		log.SetOutput(&loggerWriter{l: l.Logger(opt.StdLevel)})
	}

	return &Logs{logs: l}, nil
}

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
//
// 这是一个非必须的方法，调用可能会有一定的性能提升。
func Destroy(l *ParamsLogs) { paramsLogsPool.Put(l) }
