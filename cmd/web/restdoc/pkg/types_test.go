// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"go/types"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

var _ typeList = &types.TypeList{}

func TestPackages_TypeOf_basic_type(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)

	p.ScanDir(context.Background(), "./testdir", true)

	typ, err := p.TypeOf(context.Background(), "uint")
	a.NotError(err).
		Equal(typ, types.Typ[types.Uint])

	typ, err = p.TypeOf(context.Background(), "github.com/issue9/web/restdoc/pkg.Int")
	a.NotError(err)
	named, ok := typ.(*Named)
	a.True(ok).Equal(named.Doc().Text(), "INT\n").
		Equal(named.Next(), types.Typ[types.Int])

	typ, err = p.TypeOf(context.Background(), "github.com/issue9/web/restdoc/pkg.X")
	a.NotError(err)
	named, ok = typ.(*Named) // x = unit32
	a.True(ok).Equal(named.Doc().Text(), "X\n")
	typ = named.Next()
	a.NotNil(typ)
	named, ok = typ.(*Named) // uint32 = int8
	a.True(ok, "%T", typ).Equal(named.Doc().Text(), "").
		Equal(named.Next(), types.Typ[types.Uint8])

	pkgPath := "github.com/issue9/web/restdoc.NotExists"
	typ, err = p.TypeOf(context.Background(), pkgPath)
	a.NotError(err).Equal(typ, NotFound(pkgPath))

	typ, err = p.TypeOf(context.Background(), "invalid")
	a.Equal(err, web.NewLocaleError("%s is not a valid basic type", "invalid")).
		Nil(typ)

	typ, err = p.TypeOf(context.Background(), "{}")
	a.NotError(err).Nil(typ)

	typ, err = p.TypeOf(context.Background(), "github.com/issue9/web/restdoc.{}")
	a.NotError(err).Nil(typ)

	pkgPath = "github.com/issue9/web/restdoc{}"
	typ, err = p.TypeOf(context.Background(), pkgPath)
	a.NotError(err).Equal(typ, pkgPath)

	typ, err = p.TypeOf(context.Background(), "map")
	a.NotError(err).Equal(typ, types.NewInterfaceType(nil, nil))
}

func TestPackages_TypeOf(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", true)

	eq := func(a *assert.Assertion, path string, docs ...string) {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)

		for _, doc := range docs {
			named, ok := typ.(*Named)
			a.True(ok, "not Named:%T:%+v", typ, typ).Equal(named.Doc().Text(), doc)
			typ = named.Next()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "%+T:%+v", typ, typ).
			Equal(st.FieldDoc(1).Text(), "INT\n").
			Equal(st.Field(1).Name(), "F1").
			Equal(st.FieldDoc(3).Text(), "F2 Doc\n").
			Equal(st.Field(3).Name(), "F2")
	}

	t.Run("pkg.S", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg.S", "S Doc\n")
	})

	t.Run("pkg.S2", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg.S2", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.S2", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.S2", "", "S Doc\n")
	})

	t.Run("pkg/testdir2.S3", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.S3", "", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.S4", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.S4", "S4\n", "", "S Doc\n")
	})

	t.Run("pkg/testdir2.S5", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.S5", "S5\n", "", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.S6", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.S6", "", "S4\n", "", "S Doc\n")
	})

	t.Run("pkg/testdir2.S7", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.S7", "", "S5\n", "", "S2 Alias\n", "S Doc\n")
	})
}

func TestPackages_TypeOf_slice(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", true)

	eq := func(a *assert.Assertion, path string, arr bool, docs ...string) {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)
		named, ok := typ.(*Named)
		a.True(ok, "named error:%T", typ).NotNil(named).Equal(named.Doc().Text(), docs[0])

		if arr {
			typ = named.Next()
			array, ok := typ.(*types.Array)
			a.True(ok, "array error:%T", typ).NotNil(array)

			named, ok = array.Elem().(*Named)
			a.True(ok, "array elem error:%T", array.Elem).Equal(named.Doc().Text(), docs[1])
			typ = named.Next()
		} else {
			typ = named.Next()
			slice, ok := typ.(*types.Slice)
			a.True(ok, "slice error:%T", typ).NotNil(slice)

			named, ok = slice.Elem().(*Named)
			a.True(ok, "slice elem error:%T", slice.Elem).Equal(named.Doc().Text(), docs[1])
			typ = named.Next()
		}

		for index, doc := range docs[2:] {
			named, ok := typ.(*Named)
			a.True(ok, "named error at %d: %T", index, typ).Equal(named.Doc().Text(), doc)
			typ = named.Next()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "struct error: %T", typ).Equal(st.FieldDoc(1).Text(), "INT\n").
			Equal(st.Field(1).Name(), "F1")
	}

	t.Run("pkg/testdir2.A1", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.A1", false, "A1 Doc\n", "", "S Doc\n")
	})

	t.Run("pkg/testdir2.A2", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.A2", false, "", "", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.A3", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.A3", false, "", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.A4", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.A4", false, "", "S Doc\n")
	})

	t.Run("pkg/testdir2.A5", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.A5", true, "", "", "S Doc\n")
	})
}

func TestPackages_TypeOf_pointer(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", true)

	eq := func(a *assert.Assertion, path string, docs ...string) {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)

		named, ok := typ.(*Named)
		a.True(ok, "named error:%T", typ).NotNil(named).Equal(named.Doc().Text(), docs[0])

		typ = named.Next()
		pointer, ok := typ.(*types.Pointer)
		a.True(ok, "pointer error:%T", typ).NotNil(pointer)

		named, ok = pointer.Elem().(*Named)
		a.True(ok, "elem named error:%T", pointer.Elem).Equal(named.Doc().Text(), docs[1])
		typ = named.Next()

		for index, doc := range docs[2:] {
			named, ok := typ.(*Named)
			a.True(ok, "named error at %d: %T", index, typ).Equal(named.Doc().Text(), doc)
			typ = named.Next()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "struct error: %T", typ).Equal(st.FieldDoc(1).Text(), "INT\n").
			Equal(st.Field(1).Name(), "F1")
	}

	t.Run("pkg/testdir2.A1", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.P1", "P1 Doc\n", "", "S Doc\n")
	})

	t.Run("pkg/testdir2.A2", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.P2", "", "", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.A3", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.P3", "", "S2 Alias\n", "S Doc\n")
	})

	t.Run("pkg/testdir2.A4", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.P4", "", "S Doc\n")
	})
}

func TestPackages_TypeOf_generic(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", true)

	eq := func(a *assert.Assertion, path string, docs ...string) *Struct {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)

		for _, doc := range docs {
			named, ok := typ.(*Named)
			a.True(ok).Equal(named.Doc().Text(), doc).NotEmpty(named.ID())
			typ = named.Next()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "struct error: %+T", typ).
			Equal(st.FieldDoc(0).Text(), "F1 Doc\n").
			Equal(st.Field(0).Name(), "F1").
			Equal(st.FieldDoc(1).Text(), "F2 Doc\n").
			Equal(st.Field(1).Name(), "F2")

		return st
	}

	// 未指定泛型，应该返回错误
	t.Run("pkg.G", func(_ *testing.T) {
		path := "github.com/issue9/web/restdoc/pkg.G"
		_, err := p.TypeOf(context.Background(), path)
		a.Equal(err, web.NewLocaleError("uninstance type %s", path+"[T any]"))
	})

	t.Run("pkg.G[int]", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg.G[int]", "G Doc\n")
	})

	t.Run("pkg.GInt", func(_ *testing.T) {
		eq(a, "github.com/issue9/web/restdoc/pkg.GInt", "", "G Doc\n")
	})

	t.Run("pkg.GSNumber", func(_ *testing.T) {
		st := eq(a, "github.com/issue9/web/restdoc/pkg.GSNumber", "GSNumber Doc\n", "GS Doc\n")
		a.Equal(st.Field(2).Name(), "F3").
			Equal(st.Field(3).Name(), "F4")
	})

	t.Run("pkg/testdir2.GSNumber", func(_ *testing.T) {
		st := eq(a, "github.com/issue9/web/restdoc/pkg/testdir2.GSNumber", "", "GS Doc\n")
		a.Equal(st.Field(2).Name(), "F3").
			Equal(st.Field(3).Name(), "F4")
	})
}