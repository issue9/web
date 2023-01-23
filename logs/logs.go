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
	"github.com/issue9/term/v3/colors"
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

type (
	Level  = logs.Level
	Writer = logs.Writer
	Logger = logs.Logger
	Entry  = logs.Entry

	// Logs 日志对象
	//
	// 底层是 [logs.Logs] 类型，如果有必要可以强制转换成 [logs.Logs] 类型进行额外的操作。
	Logs interface {
		INFO() Logger
		DEBUG() Logger
		WARN() Logger
		TRACE() Logger
		ERROR() Logger
		FATAL() Logger
		NewEntry(Level) *Entry

		// With 构建带有指定参数的日志对象
		With(Level, map[string]any) Logger
	}
)

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

// New 声明日志实例
func New(w Writer, caller, created bool) *logs.Logs {
	o := make([]logs.Option, 0, 2)
	if caller {
		o = append(o, logs.Caller)
	}
	if created {
		o = append(o, logs.Created)
	}
	return logs.New(w, o...)
}
