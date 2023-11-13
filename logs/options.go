// SPDX-License-Identifier: MIT

package logs

import (
	"io"
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/logs/v7/writers"
	"github.com/issue9/logs/v7/writers/rotate"
	"github.com/issue9/term/v3/colors"

	"github.com/issue9/web/locales"
)

// 日志的时间格式
const (
	DateMilliLayout = logs.DateMilliLayout
	DateMicroLayout = logs.DateMicroLayout
	DateNanoLayout  = logs.DateNanoLayout

	MilliLayout = logs.MilliLayout
	MicroLayout = logs.MicroLayout
	NanoLayout  = logs.NanoLayout
)

// Options 初始化日志的选项
type Options struct {
	Handler Handler

	// 是否带调用堆栈信息
	Location bool

	// 指定创建日志的时间格式，如果为空表示不需要输出时间。
	Created string

	// 允许的日志级别
	Levels []Level

	// 对于 [Logger.Error] 输入 [xerrors.Formatter] 类型时，
	// 是否输出调用堆栈信息。
	StackError bool

	// 是否接管标准库日志的输出
	//
	// 如果为 true，则在 go1.21 之前会接管 log.Default() 的输出；
	// go1.21 及之后的版本则接管 log/slog.Default() 的输出；
	// 具体参考 [logs.WithStd]。
	Std bool
}

func optionsSanitize(o *Options) (*Options, error) {
	if o == nil {
		o = &Options{}
	}

	if o.Handler == nil {
		o.Handler = NewNopHandler()
	}

	for index, lv := range o.Levels {
		if !logs.IsValidLevel(lv) {
			field := "Levels[" + strconv.Itoa(index) + "]"
			return nil, config.NewFieldError(field, locales.InvalidValue)
		}
	}

	return o, nil
}

func AllLevels() []Level { return logs.AllLevels() }

func NewTextHandler(w ...io.Writer) Handler { return logs.NewTextHandler(w...) }

func NewJSONHandler(w ...io.Writer) Handler { return logs.NewJSONHandler(w...) }

// NewTermHandler 带颜色的终端输出通道
//
// 参数说明参考 [logs.NewTermHandler]
func NewTermHandler(w io.Writer, colors map[Level]colors.Color) Handler {
	return logs.NewTermHandler(w, colors)
}

func NewNopHandler() Handler { return logs.NewNopHandler() }

// NewDispatchHandler 按不同的 [Level] 派发到不同的 [Handler] 对象
func NewDispatchHandler(d map[Level]Handler) Handler { return logs.NewDispatchHandler(d) }

// MergeHandler 合并多个 [Handler] 对象
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
