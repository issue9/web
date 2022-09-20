// SPDX-License-Identifier: MIT

package form

import (
	"encoding"
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
func marshal(v any) (url.Values, error) {
	objs := map[string]reflect.Value{}
	if err := getFields(objs, "", reflect.ValueOf(v)); err != nil {
		return nil, err
	}

	marshalByInterface := func(v any) ([]byte, error) {
		if m, ok := v.(encoding.TextMarshaler); ok {
			return m.MarshalText()
		}

		// 对象的字段，只考虑基本类型，不考虑 url.Values 等类型。

		return []byte(fmt.Sprint(v)), nil
	}

	vals := url.Values{}
	for k, v := range objs {
		if kind := v.Kind(); kind != reflect.Array && kind != reflect.Slice {
			vv, err := marshalByInterface(v.Interface())
			if err != nil {
				return nil, err
			}
			vals.Add(k, string(vv))
			continue
		}

		chkSliceType(v)
		for i := 0; i < v.Len(); i++ {
			vv, err := marshalByInterface(v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			vals.Add(k, string(vv))
		}
	}

	return vals, nil
}

// 将 form-data 数据转换到 vals 中
func unmarshal(vals url.Values, obj any) error {
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

	unmarshalByInterface := func(obj reflect.Value, val string) error {
		if m, ok := obj.Interface().(encoding.TextUnmarshaler); ok {
			return m.UnmarshalText([]byte(val))
		}
		if obj.CanAddr() {
			if m, ok := obj.Addr().Interface().(encoding.TextUnmarshaler); ok {
				return m.UnmarshalText([]byte(val))
			}
		}

		return conv.Value(val, obj)
	}

	if len(names) == 0 {
		if k := obj.Kind(); k != reflect.Slice && k != reflect.Array {
			return unmarshalByInterface(obj, val[0])
		}

		chkSliceType(obj)
		slice := obj
		for i := 0; i < len(val); i++ {
			oo := reflect.New(obj.Type().Elem())
			if err := unmarshalByInterface(oo, val[i]); err != nil {
				return err
			}
			slice = reflect.Append(slice, oo.Elem())
		}
		obj.Set(slice)
		return nil
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
	// NOTE: map 元素是 unAddressable 的

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

func getFields(kv map[string]reflect.Value, name string, rv reflect.Value) error {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Map && rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}

	switch rv.Kind() {
	case reflect.Struct:
		return getStructFields(kv, name, rv)
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return errors.New("map 类型的键值只能是字符串")
		}
		return getMapFields(kv, name, rv)
	case reflect.Chan, reflect.Func:
		return nil
	default:
		kv[name] = rv
		return nil
	}
}

func getStructFields(kv map[string]reflect.Value, parent string, rv reflect.Value) error {
	rtype := rv.Type()
	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			if err := getFields(kv, parent, rv.Field(i)); err != nil {
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
		if err := getFields(kv, name, rv.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

func getMapFields(kv map[string]reflect.Value, parent string, rv reflect.Value) error {
	for iter := rv.MapRange(); iter.Next(); {
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
