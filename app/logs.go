// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	xlogs "github.com/issue9/logs/v5"
	"github.com/issue9/term/v3/colors"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/logs"
)

var logHandlersFactory = map[string]LogsHandlerBuilder{}

// LogsHandlerBuilder 构建 [logs.Handler] 的方法
type LogsHandlerBuilder = func(args []string) (logs.Handler, func() error, error)

type logsConfig struct {
	// 是否在日志中显示调用位置
	Caller bool `xml:"caller,attr,omitempty" json:"caller,omitempty" yaml:"caller,omitempty"`

	// 是否在日志中显示时间
	Created bool `xml:"created,attr,omitempty" json:"created,omitempty" yaml:"created,omitempty"`

	// 允许开启的通道
	//
	// 为空表示所有
	Levels []logs.Level `xml:"level,omitempty" json:"levels,omitempty" yaml:"levels,omitempty"`

	// 是否接管标准库的日志
	Stdlog bool `xml:"stdlog,attr,omitempty" json:"stdlog,omitempty" yaml:"stdlog,omitempty"`

	// 日志输出对象的配置
	//
	// 为空表示 [logs.NewNopHandler] 返回的对象。
	Handlers []*logHandlerConfig `xml:"writer" json:"writers" yaml:"writers"`
}

type logHandlerConfig struct {
	// NOTE: 时间格式定义在 Args 之中，而不是作为字段在当前对象之中。
	// 一般情况下终端里，时间格式只需要显示简短的以秒作为最大单位的即可，
	// 但是在文件日志里，可能还需要带上日期时间才方便。

	// 当前 Handler 支持的通道
	//
	// 为空表示采用 [logsConfig.Levels] 的值。
	Levels []logs.Level `xml:"level,omitempty" yaml:"levels,omitempty" json:"levels,omitempty"`

	// Handler 的类型
	//
	// 可通过 [RegisterLogsHandler] 方法注册，默认包含了以下几个：
	// - file 输出至文件
	// - smtp 邮件发送的日志
	// - term 输出至终端
	Type string `xml:"type,attr" yaml:"type" json:"type"`

	// 当前日志的初始化参数
	//
	// 根据以上的 type 不同而不同：
	// - file:
	//  0: 时间格式，Go 的时间格式字符串；
	//  1: 保存目录；
	//  2: 文件格式，可以包含 Go 的时间格式化字符，以 %i 作为同名文件时的序列号；
	//  3: 文件的最大尺寸，单位 byte；
	//  4: 文件的格式，默认为 text，还可选为 json；
	// - smtp:
	//  0: 时间格式，Go 的时间格式字符串；
	//  1: 账号；
	//  2: 密码；
	//  3: 主题；
	//  4: 为 smtp 的主机地址，需要带上端口号；
	//  5: 接收邮件列表；
	//  6: 文件的格式，默认为 text，还可选为 json；
	// - term
	//  0: 时间格式，Go 的时间格式字符串；
	//  1: 输出的终端，可以是 stdout 或 stderr；
	//  2-8: Level 以及对应的字符颜色，格式：erro:blue，可用颜色：
	//   - default 默认；
	//   - black 黑；
	//   - red 红；
	//   - green 绿；
	//   - yellow 黄；
	//   - blue 蓝；
	//   - magenta 洋红；
	//   - cyan 青；
	//   - white 白；
	Args []string `xml:"arg,omitempty" yaml:"args,omitempty" json:"args,omitempty"`
}

func (conf *logsConfig) build() (*logs.Options, []func() error, *web.FieldError) {
	if conf == nil {
		return &logs.Options{}, nil, nil
	}

	if len(conf.Levels) == 0 {
		conf.Levels = logs.AllLevels()
	}

	w, c, err := conf.buildHandler()
	if err != nil {
		return nil, nil, err
	}

	return &logs.Options{
		Handler: w,
		Created: conf.Created,
		Caller:  conf.Caller,
		Levels:  conf.Levels,
		Std:     conf.Stdlog,
	}, c, nil
}

func (conf *logsConfig) buildHandler() (logs.Handler, []func() error, *web.FieldError) {
	if len(conf.Handlers) == 0 {
		return logs.NewNopHandler(), nil, nil
	}

	cleanup := make([]func() error, 0, 10)

	m := make(map[logs.Level][]logs.Handler, 6)
	for i, w := range conf.Handlers {
		field := "Writers[" + strconv.Itoa(i) + "]"

		f, found := logHandlersFactory[w.Type]
		if !found {
			return nil, nil, web.NewFieldError(field+".Type", web.NewLocaleError("%s not found", w.Type))
		}

		ww, c, err := f(w.Args)
		if err != nil {
			var ce *web.FieldError
			if errors.As(err, &ce) {
				return nil, nil, ce.AddFieldParent(field)
			}
			return nil, nil, web.NewFieldError(field+".Args", err)
		}
		if c != nil {
			cleanup = append(cleanup, c)
		}

		levels := w.Levels
		if len(levels) == 0 {
			levels = conf.Levels
		}
		for _, lv := range levels {
			m[lv] = append(m[lv], ww)
		}
	}

	d := make(map[logs.Level]logs.Handler, len(m))
	for level, ws := range m {
		d[level] = logs.MergeHandler(ws...)
	}

	return logs.NewDispatchHandler(d), cleanup, nil
}

// RegisterLogsHandler 注册日志的 [LogsWriterBuilder]
//
// name 为缓存的名称，如果存在同名，则会覆盖。
func RegisterLogsHandler(b LogsHandlerBuilder, name ...string) {
	if len(name) == 0 {
		panic("参数 name 不能为空")
	}

	for _, n := range name {
		logHandlersFactory[n] = b
	}
}

func init() {
	RegisterLogsHandler(newFileLogsHandler, "file")
	RegisterLogsHandler(newTermLogsHandler, "term")
	RegisterLogsHandler(newSMTPLogsHandler, "smtp")
}

func newFileLogsHandler(args []string) (logs.Handler, func() error, error) {
	size, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return nil, nil, err
	}

	w, err := logs.NewRotateFile(args[2], args[1], size)
	if err != nil {
		return nil, nil, err
	}

	if len(args) < 5 || args[4] == "text" {
		return logs.NewTextHandler(args[0], w), w.Close, nil
	}
	return logs.NewJSONHandler(args[0], w), w.Close, nil
}

func newSMTPLogsHandler(args []string) (logs.Handler, func() error, error) {
	sendTo := strings.Split(args[5], ",")
	w := logs.NewSMTP(args[1], args[2], args[3], args[4], sendTo)

	if len(args) < 7 || args[7] == "text" {
		return logs.NewTextHandler(args[0], w), nil, nil
	}
	return logs.NewJSONHandler(args[0], w), nil, nil
}

var colorMap = map[string]colors.Color{
	"default": colors.Default,
	"black":   colors.Black,
	"red":     colors.Red,
	"green":   colors.Green,
	"yellow":  colors.Yellow,
	"blue":    colors.Blue,
	"magenta": colors.Magenta,
	"cyan":    colors.Cyan,
	"white":   colors.White,
}

// args 参数格式如下：
// - 0 时间格式
// - 1 为输出通道，可以为 stdout 和 stderr；
// - 2-8 为 level 与颜色的配置，格式为 Info:green,Warn:yellow；
func newTermLogsHandler(args []string) (logs.Handler, func() error, error) {
	if len(args) < 2 {
		return nil, nil, web.NewFieldError("Args", locales.InvalidValue)
	}

	timeLayout := args[0]

	var w io.Writer
	switch strings.ToLower(args[1]) {
	case "stderr":
		w = os.Stderr
	case "stdout":
		w = os.Stdout
	default:
		return nil, nil, web.NewFieldError("Args[1]", locales.InvalidValue)
	}

	args = args[2:]
	if len(args) > 6 {
		return nil, nil, web.NewFieldError("Args", locales.InvalidValue)
	}
	cs := make(map[logs.Level]colors.Color, len(args))
	for index, arg := range args {
		a := strings.SplitN(arg, ":", 2)

		if len(a) != 2 || a[1] == "" {
			return nil, nil, web.NewFieldError("Args["+strconv.Itoa(2+index)+"]", locales.InvalidValue)
		}

		lv, err := xlogs.ParseLevel(a[0])
		if err != nil {
			return nil, nil, web.NewFieldError("Args["+strconv.Itoa(2+index)+"]", err)
		}

		c, found := colorMap[a[1]]
		if !found {
			return nil, nil, web.NewFieldError("Args["+strconv.Itoa(2+index)+"]", locales.InvalidValue)
		}

		cs[lv] = c
	}

	return logs.NewTermHandler(timeLayout, w, cs), nil, nil
}
