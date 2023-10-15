// SPDX-License-Identifier: MIT

// Package loggertest 提供 logger 的测试用例
package loggertest

import (
	"os"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/cmd/web/restdoc/logger"
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

	ll, err := logs.New(nil, &logs.Options{
		Levels: logs.AllLevels(),
		Handler: logs.HandlerFunc(func(r *logs.Record) {
			b := logs.NewBuffer(true)
			defer b.Free()
			b.AppendFunc(r.AppendMessage)
			s := string(b.Bytes())
			t.Records[r.Level] = append(t.Records[r.Level], s)
			os.Stderr.Write([]byte(s + "\n"))
		}),
	})
	a.NotError(err).NotNil(ll)

	t.Logger = logger.New(ll, message.NewPrinter(language.SimplifiedChinese))

	return t
}
