// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package restdoc 生成 RESTful api 文档
package restdoc

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cmdopt"
	"github.com/issue9/source"
	"github.com/issue9/web"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"

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

func TestGetRealPath(t *testing.T) {
	a := assert.New(t, false)

	path, mod, err := source.ModFile(".")
	a.NotError(err).NotNil(mod)
	dir := filepath.Dir(path)

	p1, err := getRealPath(false, mod, &modfile.Require{Mod: module.Version{Path: "github.com/issue9/web", Version: web.Version}}, dir)
	a.NotError(err).
		True(strings.HasSuffix(p1, "web@"+web.Version))

	p2, err := getRealPath(true, mod, &modfile.Require{Mod: module.Version{Path: "github.com/issue9/web", Version: web.Version}}, dir)
	a.NotError(err).NotEqual(p1, p2)
}
