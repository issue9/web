// SPDX-License-Identifier: MIT

package enum

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
)

func parseFile(a *assert.Assertion, path string) *ast.File {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	a.NotError(err).NotNil(f)

	return f
}

func TestGetValues(t *testing.T) {
	a := assert.New(t, false)
	f := parseFile(a, "./testdata/testdata.go")

	vals, err := getValues(f, []string{"t1", "t3"})
	a.NotError(err).Equal(vals, map[string][]string{
		"t1": {"t1V1", "t1V2", "t1V3", "t1V4", "t1V5"},
		"t3": {"t3V1", "V2t3"},
	})
}

func TestGetValue(t *testing.T) {
	a := assert.New(t, false)
	f := parseFile(a, "./testdata/testdata.go")

	val, err := getValue(f, "t1")
	a.NotError(err).Equal(val, []string{"t1V1", "t1V2", "t1V3", "t1V4", "t1V5"})

	val, err = getValue(f, "t3")
	a.NotError(err).Equal(val, []string{"t3V1", "V2t3"})
}

func TestCheckType(t *testing.T) {
	a := assert.New(t, false)
	f := parseFile(a, "./testdata/testdata.go")

	a.NotError(checkType(f, "t1"))
	a.NotError(checkType(f, "t3"))
	a.Equal(checkType(f, "t2"), errNotAllowedType)
	a.Equal(checkType(f, "t4"), errNotAllowedType)
	a.Equal(checkType(f, "t5"), errNotAllowedType)
	a.Equal(checkType(f, "not_exists"), web.NewLocaleError("not found type %s", "not_exists"))
}

func TestDump(t *testing.T) {
	a := assert.New(t, false)

	err := dump("header", "./testdata/testdata.go", "./testdata/testdata.out", []string{"t1", "t3"})
	a.NotError(err)
}
