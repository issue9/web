// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package enum

import (
	"go/importer"
	"go/token"
	"go/types"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
)

func parseFile(a *assert.Assertion, path string) *types.Package {
	a.TB().Helper()

	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "source", nil)
	pkg, err := imp.Import(path)
	a.NotError(err).NotNil(pkg)

	return pkg
}

func TestGetValues(t *testing.T) {
	a := assert.New(t, false)
	f := parseFile(a, "testdir/testdir.go")

	vals, err := getValues(f, []string{"t1", "t3"})
	a.NotError(err).Equal(vals, map[string][]string{
		"t1": {"t1V1", "t1V2", "t1V3", "t1V4", "t1V5"},
		"t3": {"V2t3", "t3V1"},
	})
}

func TestGetValue(t *testing.T) {
	a := assert.New(t, false)
	f := parseFile(a, "testdir/testdir.go")

	val, err := GetValue(f, "t1")
	a.NotError(err).Equal(val, []string{"t1V1", "t1V2", "t1V3", "t1V4", "t1V5"})

	val, err = GetValue(f, "t3")
	a.NotError(err).Equal(val, []string{"V2t3", "t3V1"})
}

func TestCheckType(t *testing.T) {
	a := assert.New(t, false)
	f := parseFile(a, "testdir/testdir.go")

	tt, err := checkType(f, "t1")
	a.NotError(err).NotNil(tt)

	tt, err = checkType(f, "t3")
	a.NotError(err).NotNil(tt)

	tt, err = checkType(f, "t2")
	a.NotError(err).NotNil(tt)

	tt, err = checkType(f, "t5")
	a.Equal(err, ErrNotAllowedType).Nil(tt)

	tt, err = checkType(f, "not_exists")
	a.Equal(err, web.NewLocaleError("not found enum type %s", "not_exists")).Nil(tt)
}

func TestDump(t *testing.T) {
	a := assert.New(t, false)

	err := dump("header", "./testdir/testdir.go", "./testdir/testdata.out", []string{"t1", "t3"}, true, true)
	a.NotError(err)
}
