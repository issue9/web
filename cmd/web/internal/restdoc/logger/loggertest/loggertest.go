// SPDX-License-Identifier: MIT

// Package loggertest 提供 logger 的测试用例
package loggertest

import (
	"os"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Tester struct {
	*logger.Logger
	Entries map[logger.Type][]*logger.Entry
}

func New() *Tester {
	t := &Tester{
		Entries: make(map[logger.Type][]*logger.Entry, 10),
	}

	f := logger.BuildTermHandler(os.Stderr, message.NewPrinter(language.SimplifiedChinese))

	t.Logger = logger.New(func(e *logger.Entry) {
		t.Entries[e.Type] = append(t.Entries[e.Type], e)
		f(e)
	})

	return t
}
