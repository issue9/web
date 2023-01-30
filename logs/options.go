// SPDX-License-Identifier: MIT

package logs

import (
	"github.com/issue9/logs/v4"
	"golang.org/x/text/message"
)

type Options struct {
	Writer          Writer
	Caller, Created bool
	Levels          []Level
}

// New 声明日志实例
func New(opt *Options, p *message.Printer) *logs.Logs {
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

	return l
}
