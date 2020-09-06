// SPDX-License-Identifier: MIT

package form

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"unicode"

	"github.com/issue9/conv"
)

// Tag 在 struct tag 中的标签名称
const Tag = "form"

// 将 v 转换成 form-data 格式的数据
//
// NOTE: form-data 中不需要考虑 omitempty 的情况，因为无法处理数组和切片在有没有 omitempty 下的区别。
func marshal(v interface{}) (url.Values, error) {
	objs := map[string]reflect.Value{}
	if err := getFields(objs, "", reflect.ValueOf(v)); err != nil {
		return nil, err
	}

	vals := url.Values{}
	for k, v := range objs {
		kind := v.Kind()
		if kind != reflect.Array && kind != reflect.Slice {
			vals.Add(k, fmt.Sprint(v))
			continue
		}

		chkSliceType(v)
		for i := 0; i < v.Len(); i++ {
			vals.Add(k, fmt.Sprint(v.Index(i).Interface()))
		}
	}

	return vals, nil
}

// 将  form-data 数据转换到 v 中
func unmarshal(vals url.Values, obj interface{}) error {
	val := reflect.ValueOf(obj)
	for k, v := range vals {
		if err := setField(val, strings.Split(k, "."), v); err != nil {
			return err
		}
	}
	return nil
}

func setField(obj reflect.Value, names []string, val []string) error {
	for obj.Kind() == reflect.Ptr {
		if obj.IsNil() {
			obj.Set(reflect.New(obj.Type().Elem()))
		}
		obj = obj.Elem()
	}

	if len(names) == 0 {
		ok := obj.Kind()
		if ok == reflect.Slice || ok == reflect.Array {
			chkSliceType(obj)
			return conv.Value(val, obj)
		}
		return conv.Value(val[0], obj)
	}

	switch obj.Kind() {
	case reflect.Struct:
		return setStructField(obj, names, val)
	case reflect.Map:
		if obj.Type().Key().Kind() != reflect.String {
			return errors.New("map 类型的键值只能是字符串")
		}

		if obj.IsNil() {
			obj.Set(reflect.MakeMap(obj.Type()))
		}
		return setMapField(obj, names, val)
	case reflect.Func, reflect.Chan:
		return nil
	default:
		return errors.New("无法找到对应的字段")
	}
}

func setStructField(obj reflect.Value, names []string, val []string) error {
	rtype := obj.Type()
	l := obj.NumField()
	for i := 0; i < l; i++ {
		rf := rtype.Field(i)
		if rf.Anonymous {
			if err := setField(obj.Field(i), names, val); err != nil {
				return err
			}
			continue
		}

		if name := parseTag(rf); name == "-" || name != names[0] { // 名称不匹配或是 struct tag 中标记为忽略
			continue
		}

		return setField(obj.Field(i), names[1:], val)
	}

	return nil
}

func setMapField(obj reflect.Value, names []string, val []string) error {
	// NOET: map 元素是 unaddressable 的

	key := reflect.ValueOf(names[0])
	elem := reflect.New(obj.Type().Elem()).Elem()
	if e := obj.MapIndex(key); e.IsValid() {
		elem.Set(e)
	}
	if err := setField(elem, names[1:], val); err != nil {
		return err
	}
	obj.SetMapIndex(key, elem)
	return nil
}

func getFields(kv map[string]reflect.Value, name string, rval reflect.Value) error {
	for rval.Kind() == reflect.Ptr {
		if rval.IsNil() {
			rval.Set(reflect.New(rval.Type().Elem()))
		}
		rval = rval.Elem()
	}

	if rval.Kind() == reflect.Map && rval.IsNil() {
		rval.Set(reflect.MakeMap(rval.Type()))
	}

	switch rval.Kind() {
	case reflect.Struct:
		return getStructFields(kv, name, rval)
	case reflect.Map:
		if rval.Type().Key().Kind() != reflect.String {
			return errors.New("map 类型的键值只能是字符串")
		}
		return getMapFields(kv, name, rval)
	case reflect.Chan, reflect.Func:
		return nil
	default:
		kv[name] = rval
		return nil
	}
}

func getStructFields(kv map[string]reflect.Value, parent string, rval reflect.Value) error {
	rtype := rval.Type()
	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			if err := getFields(kv, parent, rval.Field(i)); err != nil {
				return err
			}
			continue
		}

		name := parseTag(field)
		if name == "-" {
			continue
		}
		if parent != "" {
			name = parent + "." + name
		}
		if err := getFields(kv, name, rval.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

func getMapFields(kv map[string]reflect.Value, parent string, rval reflect.Value) error {
	for iter := rval.MapRange(); iter.Next(); {
		name := iter.Key().String()
		if parent != "" {
			name = parent + "." + name
		}

		if err := getFields(kv, name, iter.Value()); err != nil {
			return err
		}
	}
	return nil
}

func parseTag(sf reflect.StructField) string {
	if unicode.IsLower(rune(sf.Name[0])) {
		return "-"
	}

	tag := strings.TrimSpace(sf.Tag.Get(Tag))
	switch tag {
	case "":
		return sf.Name
	default: // '-' 也原样返回即可
		return tag
	}
}

func chkSliceType(v reflect.Value) {
	k := v.Type().Elem().Kind()
	if (k < reflect.Bool || k > reflect.Float64) && k != reflect.String {
		msg := fmt.Sprintf("slice 和 array 的元素类型只能是介于 [reflect.Bool, reflect.Float64] 之间或是 reflect.String，当前为 %s", k)
		panic(msg)
	}
}
