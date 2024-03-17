// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/localeutil"
)

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func upper(v *string) { *v = strings.ToUpper(*v) }

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func required[T any](v T) bool { return !zero(v) }

func TestNewBuilder(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder[string](S(trimRight), V(zero[string], localeutil.Phrase("required")))
	id := " "
	name, msg := f("id", &id)()
	a.Nil(msg).Empty(name)

	id = "2"
	name, msg = f("id", &id)()
	a.Equal(localeutil.Phrase("required"), msg).Equal(name, "id")

	// TODO 执行顺序是否正确？
}
