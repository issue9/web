// SPDX-License-Identifier: MIT

package validation

import (
	"sort"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	min_2 := NewRule(min(-2), "-2")
	min_3 := NewRule(min(-3), "-3")
	max50 := NewRule(max(50), "50")
	max_4 := NewRule(max(-4), "-4")

	v := New(false).
		AddField(-100, "f1", min_2, min_3).
		AddField(100, "f2", max50, max_4)
	a.Equal(v.keys, []string{"f1", "f2"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("-2"), localeutil.Phrase("50")})

	v = New(true).
		AddField(-100, "f1", min_2, min_3).
		AddField(100, "f2", max50, max_4)
	a.Equal(v.keys, []string{"f1"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("-2")})
}

func TestValidation_AddField(t *testing.T) {
	a := assert.New(t, false)

	min18 := NewRule(min(18), "不能小于 18")
	min5 := NewRule(min(5), "min-5")

	obj := &object{}
	v := New(false).
		AddField(obj.Age, "obj/age", min18)
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("不能小于 18")})

	// object
	r := root2{}
	v = New(false)
	v.AddField(r.O1, "o1", NewRule(ValidateFunc(required), "o1 required")).
		AddField(r.O2, "o2", NewRule(ValidateFunc(required), "o2 required"))
	a.Equal(v.keys, []string{"o1", "o2"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("o1 required"), localeutil.Phrase("o2 required")})

	r = root2{O1: &object{}}
	v = New(false)
	v.AddField(r.O1.Age, "o1.age", min18)
	a.Equal(v.keys, []string{"o1.age"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("不能小于 18")})

	v = New(false)
	rv := root1{Root: &root2{O1: &object{}}}
	v.AddField(rv.Root.O1.Age, "root/o1/age", min18).
		AddField(rv.F1, "f1", min5)
	a.Equal(v.keys, []string{"root/o1/age", "f1"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("不能小于 18"), localeutil.Phrase("min-5")})
}

func TestValidation_AddSliceField(t *testing.T) {
	a := assert.New(t, false)

	min5 := NewRule(min(5), "min-5")

	// 将数组当普通元素处理
	v := New(false).
		AddField([]int{1, 2, 6}, "slice", min5)
	a.Equal(v.keys, []string{"slice"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("min-5")})

	// 普通元素指定为 slice
	v = New(false).
		AddSliceField(123456, "slice", min5)
	a.Equal(v.keys, []string{"slice"}).
		Equal(v.reasons, []localeutil.LocaleStringer{isNotSlice})

	v = New(true).
		AddSliceField(123456, "slice", min5)
	a.Equal(v.keys, []string{"slice"}).
		Equal(v.reasons, []localeutil.LocaleStringer{isNotSlice})

	// exitAtError = false
	v = New(false).
		AddSliceField([]int{1, 2, 6}, "slice", min5)
	a.Equal(v.keys, []string{"slice[0]", "slice[1]"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("min-5"), localeutil.Phrase("min-5")})

	// exitAtError = true
	v = New(true).
		AddSliceField([]int{1, 2, 6}, "slice", min5)
	a.Equal(v.keys, []string{"slice[0]"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("min-5")})
}

func TestValidation_AddMapField(t *testing.T) {
	a := assert.New(t, false)

	min5 := NewRule(min(5), "min-5")

	// 将数组当普通元素处理
	v := New(false).
		AddField([]int{1, 2, 6}, "map", min5)
	a.Equal(v.keys, []string{"map"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("min-5")})

	// 普通元素指定为 map
	v = New(false).
		AddMapField(123456, "map", min5)
	a.Equal(v.keys, []string{"map"}).
		Equal(v.reasons, []localeutil.LocaleStringer{isNotMap})

	v = New(true).
		AddMapField(123456, "map", min5)
	a.Equal(v.keys, []string{"map"}).
		Equal(v.reasons, []localeutil.LocaleStringer{isNotMap})

	// exitAtError = false
	v = New(false).
		AddMapField(map[string]int{"0": 1, "2": 2, "6": 6}, "map", min5)
	sort.Strings(v.keys) // NOTE: 排序会破坏 v.keys 和 v.reasons 之间的关联。
	a.Equal(v.keys, []string{"map[0]", "map[2]"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("min-5"), localeutil.Phrase("min-5")})

	// exitAtError = true
	v = New(true).
		AddMapField(map[string]int{"0": 1, "2": 2, "6": 6}, "map", min5)
	a.Length(v.keys, 1) // map 顺序未定，无法确定哪个元素会被称验证
}

func TestValidation_When(t *testing.T) {
	a := assert.New(t, false)

	min18 := NewRule(min(18), "不能小于 18")
	notEmpty := NewRule(ValidateFunc(required), "不能为空")

	obj := &object{}
	v := New(false).
		AddField(obj, "obj/age", min18).
		When(obj.Age > 0, func(v *Validation) {
			v.AddField(obj.Name, "obj/name", notEmpty)
		})
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("不能小于 18")})

	obj = &object{Age: 15}
	v = New(false).
		AddField(obj, "obj/age", min18).
		When(obj.Age > 0, func(v *Validation) {
			v.AddField(obj.Name, "obj/name", notEmpty)
		})
	a.Equal(v.keys, []string{"obj/age", "obj/name"}).
		Equal(v.reasons, []localeutil.LocaleStringer{localeutil.Phrase("不能小于 18"), localeutil.Phrase("不能为空")})
}

func TestValidation_Locale(t *testing.T) {
	a := assert.New(t, false)

	builder := catalog.NewBuilder()
	a.NotError(builder.SetString(language.SimplifiedChinese, "lang", "chn"))
	a.NotError(builder.SetString(language.TraditionalChinese, "lang", "cht"))
	cnp := message.NewPrinter(language.SimplifiedChinese, message.Catalog(builder))
	twp := message.NewPrinter(language.TraditionalChinese, message.Catalog(builder))

	max4 := NewRule(max(4), "lang")

	v := New(false).
		AddField(5, "obj", max4)
	a.Equal(v.reasons[0].LocaleString(cnp), "chn").
		Equal(v.reasons[0].LocaleString(twp), "cht")
}
