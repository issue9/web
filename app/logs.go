// SPDX-License-Identifier: MIT

package app

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/logs/v4/writers/rotate"
	"github.com/issue9/term/v3/colors"

	"github.com/issue9/web/server"
)

var logWritersFactory = map[string]LogsWriterBuilder{}

type LogsWriter = logs.Writer

// LogsWriterBuilder 构建 LogsWriter 的方法
type LogsWriterBuilder func(args []string) (LogsWriter, server.CleanupFunc, error)

type logsConfig struct {
	// 是否在日志中显示调用位置
	Caller bool `xml:"caller,attr,omitempty" json:"caller,omitempty" yaml:"caller,omitempty"`

	// 是否在日志中显示时间
	Created bool `xml:"created,attr,omitempty" json:"created,omitempty" yaml:"created,omitempty"`

	// 允许开启的通道
	//
	// 为空表示所有
	Levels []logs.Level `xml:"level,omitempty" json:"levels,omitempty" yaml:"levels,omitempty"`

	// 日志输出对象的配置
	//
	// 为空表示 logs.NewNopWriter 返回的对象。
	Writers []*logWritterConfig `xml:"writer" json:"writers" yaml:"writers"`
}

type logWritterConfig struct {
	// 当前 Writer 支持的通道
	//
	// 为空表示 logsConfig.Levels 的值。
	Levels []logs.Level `xml:"level,omitempty" yaml:"levels,omitempty" json:"levels,omitempty"`

	// Writer 的类型
	//
	// 可通过 RegisterLogsWriter 方法注册，默认包含了以下两个：
	// - file 输出至文件
	// - term 输出至终端
	Type string `xml:"type,attr" yaml:"type" json:"type"`

	// 当前日志的初始化参数
	//
	// 根据以一的 type 不同而不同，
	// - file:
	//  0: 时间格式，Go 的时间格式字符串；
	//  1: 保存目录；
	//  2: 文件格式，可以包含 Go 的时间格式化字符，以 %i 作为同名文件时的序列号；
	//  3: 文件的最大尺寸，单位 byte；
	// - term
	//  0: 时间格式，Go 的时间格式字符串；
	//  1: 字符的的颜色值，可以包含以下值：
	//   - default 默认；
	//   - black 黑；
	//   - red 红；
	//   - green 绿；
	//   - yellow 黄；
	//   - blue 蓝；
	//   - magenta 洋红；
	//   - cyan 青；
	//   - white 白；
	//  2: 输出的终端，可以是 stdout 或 stder；
	Args []string `xml:"arg,omitempty" yaml:"args,omitempty" json:"args,omitempty"`
}

func (conf *logsConfig) build() (*logs.Logs, []server.CleanupFunc, *ConfigError) {
	if conf == nil {
		return logs.New(logs.NewNopWriter()), nil, nil
	}

	if len(conf.Levels) == 0 {
		conf.Levels = []logs.Level{logs.LevelInfo, logs.LevelWarn, logs.LevelTrace, logs.LevelDebug, logs.LevelError, logs.LevelFatal}
	}

	o := make([]logs.Option, 0, 2)
	if conf.Created {
		o = append(o, logs.Created)
	}
	if conf.Caller {
		o = append(o, logs.Caller)
	}

	w, c, err := conf.buildWriter()
	if err != nil {
		return nil, nil, err
	}

	l := logs.New(w, o...)
	l.Enable(conf.Levels...)
	return l, c, nil
}

func (conf *logsConfig) buildWriter() (LogsWriter, []server.CleanupFunc, *ConfigError) {
	if len(conf.Writers) == 0 {
		return logs.NewNopWriter(), nil, nil
	}

	cleanup := make([]server.CleanupFunc, 0, 10)

	m := make(map[logs.Level][]LogsWriter, 6)
	for i, w := range conf.Writers {
		field := "Writers[" + strconv.Itoa(i) + "]"

		f, found := logWritersFactory[w.Type]
		if !found {
			return nil, nil, &ConfigError{Field: field + ".Type", Message: localeutil.Error("%s not found", w.Type)}
		}

		ww, c, err := f(w.Args)
		if err != nil {
			if ce, ok := err.(*ConfigError); ok {
				ce.Field = field + ce.Field
				return nil, nil, ce
			}
			return nil, nil, &ConfigError{Field: field + ".Args", Message: err}
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

	d := make(map[logs.Level]LogsWriter, len(m))
	for level, ws := range m {
		d[level] = logs.MergeWriter(ws...)
	}

	return logs.NewDispatchWriter(d), cleanup, nil
}

func RegisterLogsWriter(b LogsWriterBuilder, name ...string) {
	if len(name) == 0 {
		panic("参数 name 不能为空")
	}

	for _, n := range name {
		logWritersFactory[n] = b
	}
}

func init() {
	RegisterLogsWriter(newFileLogsWriter, "file")
	RegisterLogsWriter(newTermLogsWriter, "term")
}

func newFileLogsWriter(args []string) (LogsWriter, server.CleanupFunc, error) {
	size, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return nil, nil, err
	}

	w, err := rotate.New(args[1], args[2], size)
	if err != nil {
		return nil, nil, err
	}

	return logs.NewTextWriter(args[0], w), func() error { return w.Close() }, nil
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
// 0: 时间格式
// 1: 为颜色名称，以下值有效：
//  - default
//  - red
// 2: 为输出通道，以下值有效：
//  - stdout
//  - stderr
func newTermLogsWriter(args []string) (LogsWriter, server.CleanupFunc, error) {
	if len(args) != 3 {
		return nil, nil, &ConfigError{Field: "Args", Message: localeutil.Error("invalid value %s", args)}
	}

	c, found := colorMap[strings.ToLower(args[1])]
	if !found {
		return nil, nil, &ConfigError{Field: "Args[1]", Message: localeutil.Error("invalid value %s", args[1])}
	}

	var w io.Writer
	switch strings.ToLower(args[2]) {
	case "stderr":
		w = os.Stderr
	case "stdout":
		w = os.Stdout
	default:
		return nil, nil, &ConfigError{Field: "Args[2]", Message: localeutil.Error("invalid value %s", args[2])}
	}

	return logs.NewTermWriter(args[0], c, w), nil, nil
}
