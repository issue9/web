// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package loggertest 提供 logger 的测试用例
package loggertest

import (
	"os"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web/cmd/web/restdoc/logger"
)

type Tester struct {
	*logger.Logger
	Records map[logs.Level][]string
}

type handler struct {
	l logs.Level
	t *Tester
}

func (h *handler) Handle(r *logs.Record) {
	b := logs.NewBuffer(true)
	defer b.Free()
	b.AppendFunc(r.AppendMessage)
	s := string(b.Bytes())
	h.t.Records[h.l] = append(h.t.Records[h.l], s)
	os.Stderr.Write([]byte(s + "\n"))
}

func (h *handler) New(_ bool, l logs.Level, _ []logs.Attr) logs.Handler {
	return &handler{
		t: h.t,
		l: l,
	}
}

// New 声明用于测试的日志对象
func New(a *assert.Assertion) *Tester {
	a.TB().Helper()

	t := &Tester{
		Records: make(map[logs.Level][]string, 10),
	}

	ll := logs.New(&handler{t: t}, logs.WithLevels(logs.AllLevels()...))
	a.NotNil(ll)

	t.Logger = logger.New(ll)

	return t
}
