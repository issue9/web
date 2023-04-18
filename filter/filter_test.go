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
		f := New(localeutil.Phrase("zero"), Not(zero[string]), trimRight)("name", &v)
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
		f := New(localeutil.Phrase("zero"), Not(zero[obj1]), func(t *obj1) { t.Name = "obj1" })("name", &v)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, obj1{Name: "obj1"})

		v = obj1{}
		f = New(localeutil.Phrase("zero"), Not(zero[obj1]))("name", &v)
		name, msg = f()
		a.Equal(name, "name").
			Equal(msg, localeutil.Phrase("zero"))
	})
}

func TestNewFromRule(t *testing.T) {
	t.Run("NewFromRule(nil)", func(t *testing.T) {
		a := assert.New(t, false)

		v := obj2{}
		f := NewFromRule(nil, func(t *obj1) { t.Age = 18 })("name", &v.o2)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, obj2{o2: obj1{Age: 18}})

		v = obj2{}
		f = NewFromRule(nil, func(t **obj1) { *t = &obj1{Name: "obj1"} })("name", &v.o1)
		name, msg = f()
		a.Empty(name).
			Equal(v, obj2{o1: &obj1{Name: "obj1"}})
	})
}
