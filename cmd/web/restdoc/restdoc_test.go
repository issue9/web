// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package restdoc 生成 RESTful api 文档
package restdoc

import (
	"flag"
	"os"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cmdopt"

	"github.com/issue9/web/cmd/web/locales"
)

func TestInit(t *testing.T) {
	a := assert.New(t, false)
	opt := cmdopt.New(os.Stdout, flag.ContinueOnError, "", nil, nil)
	p, err := locales.NewPrinter("zh-CN")
	a.NotError(err).NotNil(p)
	Init(opt, p)

	path := "./parser/testdata/restdoc.out.yaml"
	err = opt.Exec([]string{"restdoc", "-o=" + path, "./parser/testdata"})
	a.NotError(err).FileExists(path)
}
