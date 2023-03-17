// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

type obj1 struct {
	Name string
	Age  int
}

type obj2 struct {
	o1 *obj1
	o2 obj1
}

func TestNew(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		a := assert.New(t, false)

		v := "str "
		f := New("name", &v, trimRight, NewRule(Not(zero[string]), localeutil.Phrase("zero")))
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, "str")

		v = ""
		name, msg = f()
		a.Equal(name, "name").
			Equal(msg, localeutil.Phrase("zero"))
	})

	t.Run("object", func(t *testing.T) {
		a := assert.New(t, false)

		v := obj1{Name: "name"}
		f := New("name", &v, func(t *obj1) { t.Name = "obj1" }, NewRule(Not(zero[obj1]), localeutil.Phrase("zero")))
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, obj1{Name: "obj1"})

		v = obj1{}
		f = New("name", &v, nil, NewRule(Not(zero[obj1]), localeutil.Phrase("zero")))
		name, msg = f()
		a.Equal(name, "name").
			Equal(msg, localeutil.Phrase("zero"))
	})

	t.Run("nil", func(t *testing.T) {
		a := assert.New(t, false)

		v := obj2{}
		f := New("name", &v.o2, func(t *obj1) { t.Age = 18 }, nil)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, obj2{o2: obj1{Age: 18}})

		v = obj2{}
		f = New("name", &v.o1, func(t **obj1) { *t = &obj1{Name: "obj1"} }, nil)
		name, msg = f()
		a.Empty(name).
			Equal(v, obj2{o1: &obj1{Name: "obj1"}})
	})
}
