// SPDX-License-Identifier: MIT

package logs

import (
	"io"
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/logs/v5"
	"github.com/issue9/logs/v5/writers"
	"github.com/issue9/logs/v5/writers/rotate"
	"github.com/issue9/term/v3/colors"

	"github.com/issue9/web/locales"
)

// 日志的时间格式
const (
	MilliLayout = logs.MilliLayout
	MicroLayout = logs.MicroLayout
	NanoLayout  = logs.NanoLayout
)

var allLevels = []Level{Info, Warn, Trace, Debug, Error, Fatal}

// Options 初始化日志的选项
type Options struct {
	Handler Handler
	Caller  bool    // 是否带调用堆栈信息
	Created bool    // 是否带时间
	Levels  []Level // 允许的日志通道

	// 是否接管标准库日志的输出
	//
	// 如果为 true，则在 go1.21 之前会接管 log.Default() 的输出；
	// go1.21 及之后的版本则接管 log/slog.Default() 的输出；
	Std bool
}

func optionsSanitize(o *Options) (*Options, error) {
	if o == nil {
		o = &Options{}
	}

	for index, lv := range o.Levels {
		if !logs.IsValidLevel(lv) {
			field := "Levels[" + strconv.Itoa(index) + "]"
			return nil, config.NewFieldError(field, locales.InvalidValue)
		}
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
