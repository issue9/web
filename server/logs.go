// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

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

// LogsHandlerBuilder 构建 [logs.Handler] 的方法
type LogsHandlerBuilder = func(args []string) (logs.Handler, func() error, error)

type logsConfig struct {
	// 是否在日志中显示调用位置
	Location bool `xml:"location,attr,omitempty" json:"location,omitempty" yaml:"location,omitempty"`

	// 日志显示的时间格式
	//
	// Go 的时间格式字符串，如果为空表示不显示；
	Created string `xml:"created,omitempty" json:"created,omitempty" yaml:"created,omitempty"`

	// 允许开启的通道
	//
	// 为空表示所有
	Levels []logs.Level `xml:"level,omitempty" json:"levels,omitempty" yaml:"levels,omitempty"`

	// 是否接管标准库的日志
	Std bool `xml:"std,attr,omitempty" json:"std,omitempty" yaml:"std,omitempty"`

	// 日志输出对象的配置
	//
	// 为空表示 [NewNopHandler] 返回的对象。
	Handlers []*logHandlerConfig `xml:"writer" json:"writers" yaml:"writers"`

	logs    *Logs
	cleanup []func() error
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
	//  0: 保存目录；
	//  1: 文件格式，可以包含 Go 的时间格式化字符，以 %i 作为同名文件时的序列号；
	//  2: 文件的最大尺寸，单位 byte；
	//  3: 文件的格式，默认为 text，还可选为 json；
	// - smtp:
	//  0: 账号；
	//  1: 密码；
	//  2: 主题；
	//  3: 为 smtp 的主机地址，需要带上端口号；
	//  4: 接收邮件列表；
	//  5: 文件的格式，默认为 text，还可选为 json；
	// - term
	//  0: 输出的终端，可以是 stdout 或 stderr；
	//  1-7: Level 以及对应的字符颜色，格式：erro:blue，可用颜色：
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

// AllLevels 返回所有的日志类型
func AllLevels() []logs.Level { return logs.AllLevels() }

// NewTextHandler 声明文本类型的日志输出通道
func NewTextHandler(w ...io.Writer) logs.Handler { return logs.NewTextHandler(w...) }

// NewJSONHandler 声明 JSON 类型的日志输出通道
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

func (conf *logsConfig) build() *web.FieldError {
	if conf.logs == nil {
		conf.logs = &Logs{}
	}

	if len(conf.Levels) == 0 { // 确保 buildHandler() 从 conf.Levels 继承的数据不是空的
		conf.Levels = AllLevels()
	}

	w, c, err := conf.buildHandler()
	if err != nil {
		return err
	}

	conf.logs = &Logs{
		Handler:  w,
		Created:  conf.Created,
		Location: conf.Location,
		Levels:   conf.Levels,
		Std:      conf.Std,
	}
	conf.cleanup = c

	return nil
}

func (conf *logsConfig) buildHandler() (logs.Handler, []func() error, *web.FieldError) {
	switch len(conf.Handlers) {
	case 0:
		return logs.NewNopHandler(), nil, nil
	case 1:
		item := conf.Handlers[0]

		f, found := logHandlersFactory.get(item.Type)
		if !found {
			return nil, nil, web.NewFieldError("handlers[0].type", locales.ErrNotFound(item.Type))
		}

		ww, c, err := f(item.Args)
		if err != nil {
			var ce *web.FieldError
			if errors.As(err, &ce) {
				return nil, nil, ce.AddFieldParent("handlers[0]")
			}
			return nil, nil, web.NewFieldError("handlers[0].args", err)
		}

		return ww, []func() error{c}, nil
	}

	cleanup := make([]func() error, 0, 10)
	m := make(map[logs.Level][]logs.Handler, 6)
	for i, w := range conf.Handlers {
		field := "handlers[" + strconv.Itoa(i) + "]"

		f, found := logHandlersFactory.get(w.Type)
		if !found {
			return nil, nil, web.NewFieldError(field+".type", locales.ErrNotFound(w.Type))
		}

		ww, c, err := f(w.Args)
		if err != nil {
			var ce *web.FieldError
			if errors.As(err, &ce) {
				return nil, nil, ce.AddFieldParent(field)
			}
			return nil, nil, web.NewFieldError(field+".args", err)
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
	for _, l := range AllLevels() {
		switch ws := m[l]; {
		case ws == nil:
			d[l] = NewNopHandler()
		case len(ws) == 1:
			d[l] = ws[0]
		default:
			d[l] = logs.MergeHandler(ws...)
		}
	}

	return logs.NewDispatchHandler(d), cleanup, nil
}

func newFileLogsHandler(args []string) (logs.Handler, func() error, error) {
	size, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, nil, err
	}

	w, err := NewRotateFile(args[1], args[0], size)
	if err != nil {
		return nil, nil, err
	}

	if len(args) < 4 || args[3] == "text" {
		return NewTextHandler(w), w.Close, nil
	}
	return NewJSONHandler(w), w.Close, nil
}

func newSMTPLogsHandler(args []string) (logs.Handler, func() error, error) {
	sendTo := strings.Split(args[4], ",")
	w := NewSMTP(args[0], args[1], args[2], args[3], sendTo)

	if len(args) < 6 || args[6] == "text" {
		return NewTextHandler(w), nil, nil
	}
	return NewJSONHandler(w), nil, nil
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
// - 0 为输出通道，可以为 stdout 和 stderr；
// - 1-7 为 level 与颜色的配置，格式为 Info:green,Warn:yellow；
func newTermLogsHandler(args []string) (logs.Handler, func() error, error) {
	if len(args) < 1 {
		return nil, nil, web.NewFieldError("args", locales.InvalidValue)
	}

	var w io.Writer
	switch strings.ToLower(args[0]) {
	case "stderr":
		w = os.Stderr
	case "stdout":
		w = os.Stdout
	default:
		return nil, nil, web.NewFieldError("args[0]", locales.InvalidValue)
	}

	args = args[1:]
	if len(args) > 5 {
		return nil, nil, web.NewFieldError("args", locales.InvalidValue)
	}
	cs := make(map[logs.Level]colors.Color, len(args))
	for index, arg := range args {
		a := strings.SplitN(arg, ":", 2)

		if len(a) != 2 || a[1] == "" {
			return nil, nil, web.NewFieldError("args["+strconv.Itoa(1+index)+"]", locales.InvalidValue)
		}

		lv, err := logs.ParseLevel(a[0])
		if err != nil {
			return nil, nil, web.NewFieldError("args["+strconv.Itoa(1+index)+"]", err)
		}

		c, found := colorMap[a[1]]
		if !found {
			return nil, nil, web.NewFieldError("args["+strconv.Itoa(1+index)+"]", locales.InvalidValue)
		}

		cs[lv] = c
	}

	return NewTermHandler(w, cs), nil, nil
}
