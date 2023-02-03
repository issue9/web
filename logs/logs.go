// SPDX-License-Identifier: MIT

// Package logs 日志操作
//
// 这是对 [logs] 的二次定义，方便使用者引用。
//
// [logs]: https://github.com/issue9/logs
package logs

import (
	"io"

	"github.com/issue9/logs/v4"
	"github.com/issue9/logs/v4/writers/rotate"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/message"
)

// 日志的时间格式
const (
	MicroLayout = logs.MicroLayout
	MilliLayout = logs.MilliLayout
	NanoLayout  = logs.NanoLayout
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

var allLevels = []Level{Info, Warn, Trace, Debug, Error, Fatal}

type (
	Level  = logs.Level
	Writer = logs.Writer
	Logger = logs.Logger
	Entry  = logs.Entry

	// Options 初始化日志的选项
	Options struct {
		Writer          Writer
		Caller, Created bool

		// 允许的日志通道
		Levels []Level
	}
)

func AllLevels() []Level { return allLevels }

func NewNopWriter() Writer { return logs.NewNopWriter() }

func NewTextWriter(timeLayout string, w ...io.Writer) Writer {
	return logs.NewTextWriter(timeLayout, w...)
}

// NewTermWriter 带颜色的终端输出通道
//
// timeLayout 表示输出的时间格式，遵守 time.Format 的参数要求，
// 如果为空，则不输出时间信息；
// fore 表示终端信息的字符颜色，背景始终是默认色；
// w 表示终端的接口，可以是 [os.Stderr] 或是 [os.Stdout]，
// 如果是其它的实现者则会带控制字符一起输出；
func NewTermWriter(timeLayout string, fore colors.Color, w io.Writer) Writer {
	return logs.NewTermWriter(timeLayout, fore, w)
}

func NewDispatchWriter(d map[Level]Writer) Writer { return logs.NewDispatchWriter(d) }

// MergeWriter 将多个 [Writer] 合并成一个 [Writer] 接口对象
func MergeWriter(w ...Writer) Writer { return logs.MergeWriter(w...) }

// NewRotateFile 按大小分割的文件日志
//
// 参数说明参考 [rotate.New]
func NewRotateFile(format, dir string, size int64) (*rotate.Rotate, error) {
	return rotate.New(format, dir, size)
}

// New 声明日志实例
func New(opt *Options, p *message.Printer) *Logs {
	if opt == nil {
		opt = &Options{}
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

	l := logs.New(opt.Writer, o...)
	l.Enable(opt.Levels...)

	return &Logs{logs: l}
}
