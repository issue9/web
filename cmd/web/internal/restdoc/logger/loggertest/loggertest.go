// SPDX-License-Identifier: MIT

// Package loggertest 提供 logger 的测试用例
package loggertest

import (
	"os"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
)

type Tester struct {
	*logger.Logger
	Records map[logs.Level][]string
}

func New(a *assert.Assertion) *Tester {
	a.TB().Helper()

	t := &Tester{
		Records: make(map[logs.Level][]string, 10),
	}

	ll, err := logs.New(&logs.Options{
		Levels: logs.AllLevels(),
		Handler: logs.HandleFunc(func(r *logs.Record) {
			t.Records[r.Level] = append(t.Records[r.Level], r.Message)
			os.Stderr.Write([]byte(r.Message + "\n"))
		}),
	})
	a.NotError(err).NotNil(ll)

	t.Logger = logger.New(ll, message.NewPrinter(language.SimplifiedChinese))

	return t
}
