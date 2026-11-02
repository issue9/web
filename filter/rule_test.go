// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
)

func buildMinValidator(v int) func(int) bool { return func(a int) bool { return a >= v } }

func buildMaxValidator(v int) func(int) bool { return func(a int) bool { return a < v } }

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func upper(v *string) { *v = strings.ToUpper(*v) }

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func required[T any](v T) bool { return !zero(v) }

func newFilter(exitOnError bool) (*Filter, Map) {
	m, c := CreateMapCollector()
	return New("", exitOnError, c), m
}

func TestRules(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false)
	v.Add("id", &id, Rules(Sanitizer(trimRight), Sanitizer(upper)))
	a.Length(c, 0).Equal(id, "X")
}

func TestSanitizer(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false)
	v.Add("id", &id, Sanitizer(trimRight), Sanitizer(upper))
	a.Length(c, 0).Equal(id, "X")
}

func TestValidator(t *testing.T) {
	a := assert.New(t, false)

	id := " "
	v, c := newFilter(false)
	v.Add("id", &id, Validator[string](required, localeutil.StringPhrase("required")))
	a.Empty(c).Equal(id, " ")

	id = "X "
	v, c = newFilter(false)
	v.Add("id", &id, Sanitizer(upper), Validator[string](required, localeutil.StringPhrase("required")))
	a.Empty(c).Equal(id, "X ")
}

func TestSliceSanitizer(t *testing.T) {
	a := assert.New(t, false)

	vals := []string{"s1 ", "s2"}
	v, c := newFilter(false)
	v.Add("vals", &vals, SliceSanitizer[[]string](trimRight, upper))
	a.Empty(c).Equal(vals, []string{"S1", "S2"})
}

func TestMapSanitizer(t *testing.T) {
	a := assert.New(t, false)

	vals := map[string]string{"s1 ": "s1 ", "s2": "s2"}
	v, c := newFilter(false)
	v.Add("vals", &vals, MapSanitizer[map[string]string](upper))
	a.Empty(c).Equal(vals, map[string]string{"s1 ": "S1 ", "s2": "S2"})
}

func TestSliceValidator(t *testing.T) {
	a := assert.New(t, false)

	vals := []string{"s1 ", "s2"}
	v, c := newFilter(false)
	v.Add("vals", &vals, SliceValidator[[]string](required, localeutil.StringPhrase("required")))
	a.Empty(c).Equal(vals, []string{"s1 ", "s2"})

	vals = []string{"s1 ", ""}
	v, c = newFilter(false)
	v.Add("vals", &vals, SliceValidator[[]string](required, localeutil.StringPhrase("required")))
	a.Equal(c["vals[1]"], []localeutil.StringPhrase{localeutil.StringPhrase("required")}).
		Equal(vals, []string{"s1 ", ""})
}

func TestMapValidator(t *testing.T) {
	a := assert.New(t, false)

	vals := map[string]string{"s1 ": "s1", "s2": "s2"}
	v, c := newFilter(false)
	v.Add("vals", &vals, MapValidator[map[string]string](required, localeutil.StringPhrase("required")))
	a.Empty(c)

	vals = map[string]string{"s1 ": "x", "s2": ""}
	v, c = newFilter(false)
	v.Add("vals", &vals, MapValidator[map[string]string](required, localeutil.StringPhrase("required")))
	a.Equal(c["vals[s2]"], []localeutil.StringPhrase{localeutil.StringPhrase("required")}).
		Equal(vals, map[string]string{"s1 ": "x", "s2": ""})
}
