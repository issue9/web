// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package form

import (
	"fmt"
	"net/url"
	"reflect"
	"unicode"

	"github.com/issue9/conv"
)

// Tag 在 struct tag 中的标签名称
const Tag = "form"

// 将 v 转换成 form-data 格式的数据
func marshal(v interface{}) (url.Values, error) {
	objs := map[string]reflect.Value{}
	if err := getObjectMap(objs, reflect.ValueOf(v)); err != nil {
		return nil, err
	}

	vals := url.Values{}
	for k, v := range objs {
		kind := v.Kind()
		if kind != reflect.Array && kind != reflect.Slice {
			vals.Add(k, fmt.Sprint(v))
			continue
		}

		for i := 0; i < v.Len(); i++ {
			vals.Add(k, fmt.Sprint(v.Index(i).Interface()))
		}
	}

	return vals, nil
}

// 将  form-data 数据转换到 v 中
func unmarshal(vals url.Values, obj interface{}) error {
	objs := map[string]reflect.Value{}
	if err := getObjectMap(objs, reflect.ValueOf(obj)); err != nil {
		return err
	}

	for k, v := range vals {
		field, found := objs[k]
		if !found { // 忽略未定义的字段
			continue
		}

		kind := field.Kind()
		if kind != reflect.Array && kind != reflect.Slice {
			if err := conv.Value(v[0], field); err != nil {
				return err
			}
			continue
		}

		if err := conv.Value(v, field); err != nil {
			return err
		}
	}

	return nil
}

func getObjectMap(kv map[string]reflect.Value, rval reflect.Value) error {
	for rval.Kind() == reflect.Ptr {
		if rval.IsNil() {
			rval.Set(reflect.New(rval.Type().Elem()))
		} else {
			rval = rval.Elem()
		}
	}

	rtype := rval.Type()
	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			if err := getObjectMap(kv, rval.Field(i)); err != nil {
				return err
			}
			continue
		}

		kind := field.Type.Kind()

		if kind == reflect.Func {
			continue
		}

		if kind == reflect.Map ||
			kind == reflect.Chan ||
			kind == reflect.Struct ||
			kind == reflect.Ptr {
			panic("无效的类型")
		}

		if unicode.IsLower(rune(field.Name[0])) { // 忽略以小写字母开头的字段
			continue
		}

		tag := field.Tag.Get(Tag)
		if tag == "-" {
			continue
		}

		name := field.Name
		if tag != "" {
			name = tag
		}

		kv[name] = rval.Field(i)
	}

	return nil
}
