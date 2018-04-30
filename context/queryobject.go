// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/issue9/conv"
)

const queryTag = "query"

// QueryValidator 验证接口
type QueryValidator interface {
	QueryValid(map[string]string)
}

// QueryObject 将查询参数解析到一个对象中。
//
// 返回的是每一个字段对应的错误信息。
//
//  `query:"name,default-value"`
func (ctx *Context) QueryObject(v interface{}) (errors map[string]string) {
	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	ret := make(map[string]string, rval.NumField())
	parseField(ctx.Request, rval, ret)

	// 接口在转换完成之后调用。
	if v, ok := rval.Interface().(QueryValidator); ok {
		v.QueryValid(ret)
	}

	return ret
}

func parseField(r *http.Request, rval reflect.Value, errors map[string]string) {
	rtype := rval.Type()

LOOP:
	for i := 0; i < rtype.NumField(); i++ {
		tf := rtype.Field(i)

		if tf.Anonymous {
			parseField(r, rval.Field(i), errors)
			continue
		}

		name, def := getQueryTag(tf)
		val := r.FormValue(name)
		vf := rval.Field(i)

		switch tf.Type.Kind() {
		case reflect.Slice:
			if val == "" {
				if vf.Len() > 0 { // 有默认值，则采用默认值
					continue LOOP
				}
				val = def
			}

			vals := strings.Split(val, ",")
			if len(vals) > 0 { // 指定了参数，则舍弃 slice 中的旧值
				vf.Set(vf.Slice(0, 0))
			}

			elemtype := tf.Type.Elem()
			for elemtype.Kind() == reflect.Ptr {
				elemtype = elemtype.Elem()
			}
			for _, v := range vals {
				elem := reflect.New(elemtype)
				if err := conv.Value(v, elem); err != nil {
					errors[name] = err.Error()
					continue LOOP
				}
				vf.Set(reflect.Append(vf, elem.Elem()))
			}
		case reflect.Array:
			if val == "" {
				val = def
			}
			vals := strings.Split(val, ",")

			elemtype := tf.Type.Elem()
			for elemtype.Kind() == reflect.Ptr {
				elemtype = elemtype.Elem()
			}

			if tf.Type.Len() < len(vals) { // array 类型的长度从其 type 上获取
				vals = vals[:tf.Type.Len()]
			}

			for index, v := range vals {
				elem := vf.Index(index)
				if err := conv.Value(v, elem); err != nil {
					errors[name] = err.Error()
					continue LOOP
				}
			}
		default:
			if val == "" {
				if vf.Interface() != reflect.Zero(tf.Type).Interface() {
					continue LOOP
				}
				val = def
			}

			if err := conv.Value(val, vf); err != nil {
				errors[name] = err.Error()
				continue LOOP
			}
		}
	} // end for
}

func getQueryTag(field reflect.StructField) (name, def string) {
	tag := field.Tag.Get(queryTag)
	if tag == "-" {
		return "", ""
	}

	tags := strings.SplitN(tag, ",", 2)

	switch len(tags) {
	case 0: // 都采用默认值
	case 1:
		name = strings.TrimSpace(tags[0])
	case 2:
		name = strings.TrimSpace(tags[0])
		def = strings.TrimSpace(tags[1])
	}

	if name == "" {
		name = field.Name
	}

	return name, def
}
