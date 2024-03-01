// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"go/types"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

func TestSplitFieldTypes(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	a.NotNil(p)
	p.ScanDir(context.Background(), "./testdir", true)

	path, ts, err := p.splitFieldTypes(context.Background(), "<f1=int>")
	a.NotError(err).
		Empty(path).
		Equal(ts, map[string]types.Type{"f1": types.Typ[types.Int]})

	path, ts, err = p.splitFieldTypes(context.Background(), "t1<f1=int>")
	a.NotError(err).
		Equal(path, "t1").
		Equal(ts, map[string]types.Type{"f1": types.Typ[types.Int]})

	path, ts, err = p.splitFieldTypes(context.Background(), "t1<f1=int,f2=uint8>")
	a.NotError(err).
		Equal(path, "t1").
		Equal(ts, map[string]types.Type{"f1": types.Typ[types.Int], "f2": types.Typ[types.Uint8]})

	path, ts, err = p.splitFieldTypes(context.Background(), "t1<f1=int,f2=github.com/issue9/web/restdoc/pkg.S<S=int>>")
	a.NotError(err).Equal(path, "t1")
	named, ok := ts["f2"].(*Named)
	a.True(ok, "%+T", ts["f2"])
	st, ok := named.Next().(*Struct)
	a.True(ok).NotNil(st).Equal(st.Field(2).Type(), types.Typ[types.Int])
}

func TestSplitTypes(t *testing.T) {
	a := assert.New(t, false)
	typ := types.Typ[types.Bool]

	f, path := splitTypes("bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), typ)

	f, path = splitTypes("[]bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewSlice(typ))

	f, path = splitTypes("[]*bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewSlice(types.NewPointer(typ)))

	f, path = splitTypes("[5]bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewArray(typ, 5))

	f, path = splitTypes("[\t]bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewSlice(typ))

	f, path = splitTypes("[\t11]bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewArray(typ, 11))

	f, path = splitTypes("[\t-11]bool")
	a.NotNil(f).Equal(path, "[\t-11]bool").
		Equal(f(typ), typ)

	f, path = splitTypes("[abc]bool")
	a.NotNil(f).Equal(path, "[abc]bool").
		Equal(f(typ), typ)

	f, path = splitTypes("[][]bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewSlice(types.NewSlice(typ)))

	f, path = splitTypes("[][5]*bool")
	a.NotNil(f).Equal(path, "bool").
		Equal(f(typ), types.NewSlice(types.NewArray(types.NewPointer(typ), 5)))
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
