// SPDX-License-Identifier: MIT

// Package restdoc 生成 RESTful api 文档
package restdoc

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/source"
	"github.com/issue9/web"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestGetRealPath(t *testing.T) {
	a := assert.New(t, false)

	path, mod, err := source.ModFile(".")
	a.NotError(err).NotNil(mod)
	dir := filepath.Dir(path)

	p1, err := getRealPath(false, mod, &modfile.Require{Mod: module.Version{Path: "github.com/issue9/web", Version: web.Version}}, dir)
	a.NotError(err).
		True(strings.HasSuffix(p1, "github.com/issue9/web@"+web.Version))

	p2, err := getRealPath(true, mod, &modfile.Require{Mod: module.Version{Path: "github.com/issue9/web", Version: web.Version}}, dir)
	a.NotError(err).
		NotEqual(p1, p2)
}
