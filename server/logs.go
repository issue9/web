// SPDX-License-Identifier: MIT

package server

import (
	"io"
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
	"github.com/issue9/logs/v7/writers"
	"github.com/issue9/logs/v7/writers/rotate"
	"github.com/issue9/term/v3/colors"

	"github.com/issue9/web"
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

// Logs 初始化日志的选项
type Logs struct {
	// Handler 后端处理接口
	//
	// 内置了以下几种方式：
	//  - [NewNopHandler]
	//  - [NewTermHandler]
	//  - [NewTextHandler]
	//  - [NewJSONHandler]
	Handler logs.Handler

	// 是否带调用堆栈信息
	Location bool

	// 指定创建日志的时间格式，如果为空表示不需要输出时间。
	Created string

	// 允许的日志级别
	Levels []logs.Level

	// 对于 [Logger.Error] 输入 [xerrors.Formatter] 类型时，
	// 是否输出调用堆栈信息。
	StackError bool

	// 是否接管标准库日志的输出
	Std bool
}

func (o *Options) buildLogs(p *localeutil.Printer) *web.FieldError {
	if o.Logs == nil {
		o.Logs = &Logs{}
	}

	if o.Logs.Handler == nil {
		o.Logs.Handler = NewNopHandler()
	}

	for index, lv := range o.Logs.Levels {
		if !logs.IsValidLevel(lv) {
			field := "Logs.Levels[" + strconv.Itoa(index) + "]"
			return config.NewFieldError(field, locales.InvalidValue)
		}
	}

	oo := make([]logs.Option, 0, 5)

	oo = append(oo, logs.WithLocale(p))

	if o.Logs.Location {
		oo = append(oo, logs.WithLocation(true))
	}
	if o.Logs.Created != "" {
		oo = append(oo, logs.WithCreated(o.Logs.Created))
	}
	if o.Logs.StackError {
		oo = append(oo, logs.WithDetail(true))
	}

	if o.Logs.Std {
		oo = append(oo, logs.WithStd())
	}

	if len(o.Logs.Levels) > 0 {
		oo = append(oo, logs.WithLevels(o.Logs.Levels...))
	}

	o.logs = logs.New(o.Logs.Handler, oo...)

	return nil
}

func AllLevels() []logs.Level { return logs.AllLevels() }

func NewTextHandler(w ...io.Writer) logs.Handler { return logs.NewTextHandler(w...) }

func NewJSONHandler(w ...io.Writer) logs.Handler { return logs.NewJSONHandler(w...) }

// NewTermHandler 带颜色的终端输出通道
//
// 参数说明参考 [logs.NewTermHandler]
func NewTermHandler(w io.Writer, colors map[logs.Level]colors.Color) logs.Handler {
	return logs.NewTermHandler(w, colors)
}

func NewNopHandler() logs.Handler { return logs.NewNopHandler() }

// NewDispatchHandler 按不同的 [Level] 派发到不同的 [Handler] 对象
func NewDispatchHandler(d map[logs.Level]logs.Handler) logs.Handler {
	return logs.NewDispatchHandler(d)
}

// MergeHandler 合并多个 [Handler] 对象
func MergeHandler(w ...logs.Handler) logs.Handler { return logs.MergeHandler(w...) }

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
