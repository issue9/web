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
	"github.com/issue9/logs/v5"
	"github.com/issue9/logs/v5/writers"
	"github.com/issue9/logs/v5/writers/rotate"
	"github.com/issue9/term/v3/colors"
)

// 日志的时间格式
const (
	MilliLayout = logs.MilliLayout
	MicroLayout = logs.MicroLayout
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
	Handler    = logs.Handler
	HandleFunc = logs.HandleFunc
	Logger     = logs.Logger
	Record     = logs.Record

	// Options 初始化日志的选项
	Options struct {
		Handler Handler
		Caller  bool    // 是否带调用堆栈信息
		Created bool    // 是否带时间
		Levels  []Level // 允许的日志通道

		// 标准库的错误日志重定义至哪个通道
		//
		// 一些由 log.Println 等全局方法输出的内容，由此指定输出的通道。
		StdLevel Level
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

func NewNopHandler() Handler { return logs.NewNopHandler() }

func NewTextHandler(timeLayout string, w ...io.Writer) Handler {
	return logs.NewTextHandler(timeLayout, w...)
}

func NewJSONHandler(timeLayout string, w ...io.Writer) Handler {
	return logs.NewJSONHandler(timeLayout, w...)
}

// NewTermHandler 带颜色的终端输出通道
//
// 参数说明参考 [logs.NewTermHandler]
func NewTermHandler(timeLayout string, w io.Writer, colors map[Level]colors.Color) Handler {
	return logs.NewTermHandler(timeLayout, w, colors)
}

func NewDispatchHandler(d map[Level]Handler) Handler { return logs.NewDispatchHandler(d) }

func MergeHandler(w ...Handler) Handler { return logs.MergeHandler(w...) }

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
