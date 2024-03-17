// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/localeutil"
)

func TestS(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(S(trimRight), S(upper))
	id := "x "
	name, msg := f("id", &id)()
	a.Nil(msg).Empty(name).Equal(id, "X")
}

func TestV(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(V[string](required, localeutil.Phrase("required")))
	id := " "
	name, msg := f("id", &id)()
	a.Nil(msg).Empty(name).Equal(id, " ")

	f = NewBuilder(S(upper), V[string](required, localeutil.Phrase("required")))
	id = "X "
	name, msg = f("id", &id)()
	a.Nil(msg).Empty(name).Equal(id, "X ")
}

func TestSS(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(SS[[]string](trimRight, upper))
	vals := []string{"s1 ", "s2"}
	name, msg := f("vals", &vals)()
	a.Nil(msg).Empty(name).Equal(vals, []string{"S1", "S2"})
}

func TestMS(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(MS[map[string]string](upper))
	vals := map[string]string{"s1 ": "s1 ", "s2": "s2"}
	name, msg := f("vals", &vals)()
	a.Nil(msg).Empty(name).Equal(vals, map[string]string{"s1 ": "S1 ", "s2": "S2"})
}

func TestSV(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(SV[[]string](required, localeutil.Phrase("required")))
	vals := []string{"s1 ", "s2"}
	name, msg := f("vals", &vals)()
	a.Nil(msg).Empty(name).Equal(vals, []string{"s1 ", "s2"})

	f = NewBuilder(SV[[]string](required, localeutil.Phrase("required")))
	vals = []string{"s1 ", ""}
	name, msg = f("vals", &vals)()
	a.Equal(msg, localeutil.Phrase("required")).
		Equal(name, "vals[1]").
		Equal(vals, []string{"s1 ", ""})
}

func TestMV(t *testing.T) {
	a := assert.New(t, false)

	f := NewBuilder(MV[map[string]string](required, localeutil.Phrase("required")))
	vals := map[string]string{"s1 ": "s1", "s2": "s2"}
	name, msg := f("vals", &vals)()
	a.Nil(msg).Empty(name)

	vals = map[string]string{"s1 ": "x", "s2": ""}
	name, msg = f("vals", &vals)()
	a.Equal(msg, localeutil.Phrase("required")).
		Equal(name, "vals[s2]").
		Equal(vals, map[string]string{"s1 ": "x", "s2": ""})
}
