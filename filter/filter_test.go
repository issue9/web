// SPDX-License-Identifier: MIT

package filter

import (
	"reflect"
	"strings"
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

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func nz[T any](v T) bool { return !reflect.ValueOf(v).IsZero() }

func between[T ~int | ~uint | float32 | float64](min, max T) func(T) bool {
	return func(vv T) bool { return vv >= min && vv <= max }
}

func TestNewFromVS(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		a := assert.New(t, false)

		v := "str "
		f := NewFromVS(localeutil.Phrase("zero"), nz[string], trimRight)("name", &v)
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
		f := NewFromVS(localeutil.Phrase("zero"), nz[obj1], func(t *obj1) { t.Name = "obj1" })("name", &v)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, obj1{Name: "obj1"})

		v = obj1{}
		f = NewFromVS(localeutil.Phrase("zero"), nz[obj1])("name", &v)
		name, msg = f()
		a.Equal(name, "name").
			Equal(msg, localeutil.Phrase("zero"))
	})
}

func TestNew(t *testing.T) {
	t.Run("NewFromRule(nil)", func(t *testing.T) {
		a := assert.New(t, false)

		v := obj2{}
		f := New(nil, func(t *obj1) { t.Age = 18 })("name", &v.o2)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, obj2{o2: obj1{Age: 18}})

		v = obj2{}
		f = New(nil, func(t **obj1) { *t = &obj1{Name: "obj1"} })("name", &v.o1)
		name, msg = f()
		a.Empty(name).Nil(msg).
			Equal(v, obj2{o1: &obj1{Name: "obj1"}})
	})
}

func TestNewSlice(t *testing.T) {
	a := assert.New(t, false)
	message := localeutil.Phrase("error")
	rule := NewSliceRule[int, []int](func(val int) bool { return val > 0 }, message)

	f := NewSlice(rule, func(t *int) { *t += 1 })
	a.NotNil(f)
	v := []int{1, 2, 3, 4, 5}
	name, msg := f("slice", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, []int{2, 3, 4, 5, 6})

	// rule == nil
	f = NewSlice[int, []int](nil, func(t *int) { *t += 1 })
	a.NotNil(f)
	v = []int{1, 2, 3, 4, 5}
	name, msg = f("slice", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, []int{2, 3, 4, 5, 6})

	f = NewSlice(rule, func(t *int) { *t -= 1 })
	a.NotNil(f)
	v = []int{1, 2, 3, 4, 5}
	name, msg = f("slice", &v)()
	a.Equal(name, "slice[0]").
		Equal(msg, localeutil.Phrase("error"))
}

func TestNewMap(t *testing.T) {
	a := assert.New(t, false)
	message := localeutil.Phrase("error")
	rule := NewMapRule[int, int, map[int]int](func(val int) bool { return val > 0 }, message)

	f := NewMap(rule, func(t *int) { *t += 1 })
	a.NotNil(f)
	v := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	name, msg := f("map", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, map[int]int{1: 2, 2: 3, 3: 4, 4: 5, 5: 6})

	// rule == nil
	f = NewMap[int, int, map[int]int](nil, func(t *int) { *t += 1 })
	a.NotNil(f)
	v = map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	name, msg = f("map", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, map[int]int{1: 2, 2: 3, 3: 4, 4: 5, 5: 6})

	f = NewMap(rule, func(t *int) { *t -= 1 })
	a.NotNil(f)
	v = map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	name, msg = f("map", &v)()
	a.Equal(name, "map[1]").
		Equal(msg, localeutil.Phrase("error"))
}
