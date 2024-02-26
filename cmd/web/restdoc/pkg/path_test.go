// SPDX-License-Identifier: MIT

package pkg

import (
	"go/types"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestSplitTypes(t *testing.T) {
	a := assert.New(t, false)
	typ := types.Typ[types.Bool]

	f, path := splitTypes("bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), typ)

	f, path = splitTypes("[]bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewSlice(typ))

	f, path = splitTypes("[]*bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewSlice(types.NewPointer(typ)))

	f, path = splitTypes("[5]bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewArray(typ, 5))

	f, path = splitTypes("[\t]bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewSlice(typ))

	f, path = splitTypes("[\t11]bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewArray(typ, 11))

	f, path = splitTypes("[\t-11]bool")
	a.NotNil(f).Equal(path, "[\t-11]bool")
	a.Equal(f(typ), typ)

	f, path = splitTypes("[abc]bool")
	a.NotNil(f).Equal(path, "[abc]bool")
	a.Equal(f(typ), typ)

	f, path = splitTypes("[][]bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewSlice(types.NewSlice(typ)))

	f, path = splitTypes("[][5]*bool")
	a.NotNil(f).Equal(path, "bool")
	a.Equal(f(typ), types.NewSlice(types.NewArray(types.NewPointer(typ), 5)))
}

func TestFilterVersionSuffix(t *testing.T) {
	a := assert.New(t, false)

	p, ok := filterVersionSuffix("github.com/issue9/logs/v6", '/')
	a.True(ok).Equal(p, "github.com/issue9/logs")

	p, ok = filterVersionSuffix("github.com/issue9/logs.v6", '.')
	a.True(ok).Equal(p, "github.com/issue9/logs")

	p, ok = filterVersionSuffix("github.com/issue9/logs", '/')
	a.False(ok).Equal(p, "github.com/issue9/logs")
}
