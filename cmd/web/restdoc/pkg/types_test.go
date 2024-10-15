// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"go/types"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

var _ typeList = &types.TypeList{}

func TestPackages_TypeOf_basic_type(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	a.NotNil(p)

	p.ScanDir(context.Background(), "./testdir", true)

	typ, err := p.TypeOf(context.Background(), "uint")
	a.NotError(err).
		Equal(typ, types.Typ[types.Uint])

	typ, err = p.TypeOf(context.Background(), "github.com/issue9/web/restdoc/pkg.Int")
	a.NotError(err)
	alias, ok := typ.(*Alias)
	a.True(ok, "%T", typ).Equal(alias.Doc().Text(), "INT\n").
		Equal(alias.Rhs(), types.Typ[types.Int])

	typ, err = p.TypeOf(context.Background(), "github.com/issue9/web/restdoc/pkg.X")
	a.NotError(err)
	alias, ok = typ.(*Alias) // x = unit32
	a.True(ok).Equal(alias.Doc().Text(), "X\n")
	typ = alias.Rhs()
	a.NotNil(typ)
	alias, ok = typ.(*Alias) // uint32 = int8
	a.True(ok, "%T", typ).Equal(alias.Doc().Text(), "").
		Equal(alias.Rhs(), types.Typ[types.Uint8])

	pkgPath := "github.com/issue9/web/restdoc.NotExists"
	typ, err = p.TypeOf(context.Background(), pkgPath)
	a.NotError(err).Equal(typ, NotFound(pkgPath))

	typ, err = p.TypeOf(context.Background(), "invalid")
	a.NotError(err).Equal(typ, NotFound("invalid"))

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

	eq := func(a *assert.Assertion, path string, docs ...string) *Struct {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)

		for _, doc := range docs {
			alias, ok := typ.(*Alias)
			a.True(ok, "not Alias:%T:%+v", typ, typ).Equal(alias.Doc().Text(), doc)
			typ = alias.Rhs()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "%+T:%+v", typ, typ).
			Equal(st.FieldDoc(0).Text(), "").
			Equal(st.Field(0).Name(), "").
			True(st.Field(0).Embedded()).
			NotNil(st.Field(0).Type()).
			Equal(st.FieldDoc(1).Text(), "INT\n").
			Equal(st.Field(1).Name(), "F1").
			Equal(st.FieldDoc(3).Text(), "F2 Doc\n").
			Equal(st.Field(3).Name(), "F2")

		return st
	}

	t.Run("pkg.S", func(_ *testing.T) {
		s := eq(a, "github.com/issue9/web/restdoc/pkg.S", "S Doc\n")
		a.Equal(s.NumFields(), 5)
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
		alias, ok := typ.(*Alias)
		a.True(ok, "alias error:%T", typ).NotNil(alias).Equal(alias.Doc().Text(), docs[0])

		if arr {
			typ = alias.Rhs()
			array, ok := typ.(*types.Array)
			a.True(ok, "array error:%T", typ).NotNil(array)

			alias, ok = array.Elem().(*Alias)
			a.True(ok, "array elem error:%T", array.Elem).Equal(alias.Doc().Text(), docs[1])
			typ = alias.Rhs()
		} else {
			typ = alias.Rhs()
			slice, ok := typ.(*types.Slice)
			a.True(ok, "slice error:%T", typ).NotNil(slice)

			alias, ok = slice.Elem().(*Alias)
			a.True(ok, "slice elem error:%T", slice.Elem).Equal(alias.Doc().Text(), docs[1])
			typ = alias.Rhs()
		}

		for index, doc := range docs[2:] {
			alias, ok := typ.(*Alias)
			a.True(ok, "alias error at %d: %T", index, typ).Equal(alias.Doc().Text(), doc)
			typ = alias.Rhs()
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

		alias, ok := typ.(*Alias)
		a.True(ok, "alias error:%T", typ).NotNil(alias).Equal(alias.Doc().Text(), docs[0])

		typ = alias.Rhs()
		pointer, ok := typ.(*types.Pointer)
		a.True(ok, "pointer error:%T", typ).NotNil(pointer)

		alias, ok = pointer.Elem().(*Alias)
		a.True(ok, "elem alias error:%T", pointer.Elem).Equal(alias.Doc().Text(), docs[1])
		typ = alias.Rhs()

		for index, doc := range docs[2:] {
			alias, ok := typ.(*Alias)
			a.True(ok, "alias error at %d: %T", index, typ).Equal(alias.Doc().Text(), doc)
			typ = alias.Rhs()
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

	eqG := func(a *assert.Assertion, path string, docs ...string) {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)

		for _, doc := range docs {
			alias, ok := typ.(*Alias)
			a.True(ok).Equal(alias.Doc().Text(), doc).NotEmpty(alias.ID())
			typ = alias.Rhs()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "struct error: %+T", typ).
			Equal(st.FieldDoc(0).Text(), "F1 Doc\n").
			Equal(st.Field(0).Name(), "F1").
			Equal(st.FieldDoc(1).Text(), "F2 Doc\n").
			Equal(st.Field(1).Name(), "F2")
	}

	// 未指定泛型，应该返回错误
	t.Run("pkg.G", func(_ *testing.T) {
		path := "github.com/issue9/web/restdoc/pkg.G"
		_, err := p.TypeOf(context.Background(), path)
		a.Equal(err, web.NewLocaleError("not found type param %s", "T"))
	})

	t.Run("pkg.G[int]", func(_ *testing.T) {
		eqG(a, "github.com/issue9/web/restdoc/pkg.G[int]", "G Doc\n")
	})

	t.Run("pkg.GInt", func(_ *testing.T) {
		eqG(a, "github.com/issue9/web/restdoc/pkg.GInt", "", "G Doc\n")
	})

	l = loggertest.New(a)
	p = New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", true)

	eqGS := func(a *assert.Assertion, path string, docs ...string) *Struct {
		a.TB().Helper()
		typ, err := p.TypeOf(context.Background(), path)
		a.NotError(err).NotNil(typ)

		for _, doc := range docs {
			alias, ok := typ.(*Alias)
			a.True(ok).Equal(alias.Doc().Text(), doc).NotEmpty(alias.ID())
			typ = alias.Rhs()
		}

		st, ok := typ.(*Struct)
		a.True(ok, "struct error: %+T", typ).
			Equal(st.FieldDoc(1).Text(), "").
			Equal(st.Field(1).Name(), "F3").
			Equal(st.Field(1).Type(), types.Typ[types.Int]).
			Equal(st.FieldDoc(2).Text(), "").
			Equal(st.Field(2).Name(), "F4").
			Equal(st.FieldDoc(3).Text(), "引用类型的字段\n").
			Equal(st.Field(3).Name(), "F5")

		return st
	}

	t.Run("pkg.GSNumber", func(_ *testing.T) {
		s := eqGS(a, "github.com/issue9/web/restdoc/pkg.GSNumber", "GSNumber Doc\n", "GS Doc\n")
		a.Equal(s.NumFields(), 6)
	})

	t.Run("pkg/testdir2.GSNumber", func(_ *testing.T) {
		s := eqGS(a, "github.com/issue9/web/restdoc/pkg/testdir2.GSNumber", "", "GS Doc\n")
		a.Equal(s.NumFields(), 5)
	})
}

func TestAlias_ID(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", true)

	path := "github.com/issue9/web/restdoc/pkg.G[int]"
	typ, err := p.TypeOf(context.Background(), path)
	a.NotError(err).NotNil(typ)
	alias, ok := typ.(*Alias)
	a.True(ok).NotNil(alias).Equal(alias.ID(), "github.com/issue9/web/restdoc/pkg.G[int]")

	path = "github.com/issue9/web/restdoc/pkg.GInt"
	typ, err = p.TypeOf(context.Background(), path)
	a.NotError(err).NotNil(typ)
	alias, ok = typ.(*Alias)
	a.True(ok).NotNil(alias).Equal(alias.ID(), "github.com/issue9/web/restdoc/pkg.GInt")

	path = "github.com/issue9/web/restdoc/pkg.GSNumber"
	typ, err = p.TypeOf(context.Background(), path)
	a.NotError(err).NotNil(typ)
	alias, ok = typ.(*Alias)
	a.True(ok).NotNil(alias).Equal(alias.ID(), "github.com/issue9/web/restdoc/pkg.GSNumber")

	// pkg.GS[...]

	path = "github.com/issue9/web/restdoc/pkg.GS[int,Int,github.com/issue9/web/restdoc/pkg.Int]"
	typ, err = p.TypeOf(context.Background(), path)
	a.NotError(err).NotNil(typ)
	alias, ok = typ.(*Alias)
	a.True(ok).NotNil(alias).Equal(alias.ID(), "github.com/issue9/web/restdoc/pkg.GS[int,Int,github.com/issue9/web/restdoc/pkg.Int]")

	s, ok := alias.Rhs().(*Struct)
	a.True(ok).NotNil(s)

	// 字段类型的 ID 是否正确
	f1, ok := s.Field(2).Type().(*Alias)
	a.True(ok, "%T", s.Field(2).Type()).NotNil(f1).Equal(f1.ID(), "github.com/issue9/web/restdoc/pkg.Int")

	// 嵌套泛型的 ID 是否正确
	f0, ok := s.Field(0).Type().(*Alias)
	a.True(ok, "%T", s.Field(0).Type()).NotNil(f0).Equal(f0.ID(), "github.com/issue9/web/restdoc/pkg.G[github.com/issue9/web/restdoc/pkg.Int]")

	// 嵌套泛型的的嵌套泛型
	s, ok = f0.Rhs().(*Struct)
	a.True(ok).NotNil(s)
	f00, ok := s.Field(0).Type().(*Alias)
	a.True(ok, "%T", s.Field(0).Type()).NotNil(f00).Equal(f00.ID(), "github.com/issue9/web/restdoc/pkg.Int")
}
