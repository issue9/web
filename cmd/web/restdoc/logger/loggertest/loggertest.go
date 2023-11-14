// SPDX-License-Identifier: MIT

// Package loggertest 提供 logger 的测试用例
package loggertest

import (
	"os"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"

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

func (h *handler) New(d bool, l logs.Level, attrs []logs.Attr) logs.Handler {
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

	ll, err := logs.New(nil, &logs.Options{
		Levels:  logs.AllLevels(),
		Handler: &handler{t: t},
	})
	a.NotError(err).NotNil(ll)

	t.Logger = logger.New(ll)

	return t
}
