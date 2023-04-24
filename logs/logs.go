// SPDX-License-Identifier: MIT

// Package logs 日志操作
//
// 这是对 [logs] 的二次定义，方便使用者引用。
//
// [logs]: https://github.com/issue9/logs
package logs

import (
	"io"
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/logs/v4/writers"
	"github.com/issue9/logs/v4/writers/rotate"
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

var allLevels = []Level{Info, Warn, Trace, Debug, Error, Fatal}

type (
	Level      = logs.Level
	Writer     = logs.Writer
	WriteEntry = logs.WriteEntry
	Logger     = logs.Logger
	Entry      = logs.Entry

	// Options 初始化日志的选项
	Options struct {
		Writer   Writer
		Caller   bool    // 是否带调用堆栈信息
		Created  bool    // 是否带时间
		Levels   []Level // 允许的日志通道
		StdLevel Level   // 标准库的错误日志重定义至哪个通道
	}
)

func optionsSanitize(o *Options) (*Options, error) {
	if o == nil {
		o = &Options{}
	}

	for index, lv := range o.Levels {
		if !logs.IsValidLevel(lv) {
			field := "Levels[" + strconv.Itoa(index) + "]"
			return nil, config.NewFieldError(field, localeutil.Phrase("invalid value"))
		}
	}

	if o.StdLevel != 0 && !logs.IsValidLevel(o.StdLevel) {
		return nil, config.NewFieldError("StdLevel", localeutil.Phrase("invalid value"))
	}

	return o, nil
}

func AllLevels() []Level { return allLevels }

func NewNopWriter() Writer { return logs.NewNopWriter() }

func NewTextWriter(timeLayout string, w ...io.Writer) Writer {
	return logs.NewTextWriter(timeLayout, w...)
}

func NewJSONWriter(timeLayout string, w ...io.Writer) Writer {
	return logs.NewJSONWriter(timeLayout, w...)
}

// NewTermWriter 带颜色的终端输出通道
//
// 参数说明参考 [logs.NewTermWriter]
func NewTermWriter(timeLayout string, fore colors.Color, w io.Writer) Writer {
	return logs.NewTermWriter(timeLayout, fore, w)
}

func NewDispatchWriter(d map[Level]Writer) Writer { return logs.NewDispatchWriter(d) }

// MergeWriter 将多个 [Writer] 合并成一个 [Writer] 接口对象
func MergeWriter(w ...Writer) Writer { return logs.MergeWriter(w...) }

// NewRotateFile 按大小分割的文件日志
//
// 参数说明参考 [rotate.New]
func NewRotateFile(format, dir string, size int64) (io.WriteCloser, error) {
	return rotate.New(format, dir, size)
}

// NewSMTP 将日志内容发送至指定邮箱
//
// 参数说明参考 [writers.NewSMTP]
func NewSMTP(username, password, subject, host string, sendTo []string) io.Writer {
	return writers.NewSMTP(username, password, subject, host, sendTo)
}
