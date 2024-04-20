// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
)

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func upper(v *string) { *v = strings.ToUpper(*v) }

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func required[T any](v T) bool { return !zero(v) }

func TestNewBuilder(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(S(trimRight), V(zero[string], localeutil.Phrase("required")))
	id := " "
	name, msg := f("id", &id)()
	a.Nil(msg).Empty(name)

	id = "2"
	name, msg = f("id", &id)()
	a.Equal(localeutil.Phrase("required"), msg).Equal(name, "id")

	// 执行顺序是否正常

	f = NewBuilder(S(func(v *string) { *v = *v + "1" }), S(func(v *string) { *v = *v + "2" }))
	id = " "
	name, msg = f("id", &id)()
	a.Nil(msg).Empty(name).Equal(id, " 12")
}

func TestToFieldError(t *testing.T) {
	a := assert.New(t, false)

	f1 := NewBuilder(S(trimRight), V(zero[string], localeutil.Phrase("required")))
	v1 := " x"
	fields := ToFieldError(f1("id", &v1))
	a.Equal(v1, " x").
		Equal(config.NewFieldError("id", localeutil.Phrase("required")), fields)

	v1 = " "
	f2 := NewBuilder(V(zero[string], localeutil.Phrase("required")))
	v2 := " x"
	fields = ToFieldError(f1("v1", &v1), f2("v2", &v2))
	a.Equal(v1, "").
		Equal(v2, " x").
		Equal(config.NewFieldError("v2", localeutil.Phrase("required")), fields)
}
