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

// QuerySanitizer 表示对一个查询参数构成的结构体进行数据验证和内容修正的接口
type QuerySanitizer interface {
	// 参数 errors 用来保存由函数中发现的错误信息。
	//
	// 其中的键名为错误字段名称，键值为错误信息。
	QuerySanitize(errors map[string]string)
}

// UnmarshalQueryer 该接口实现在了将一些特定的查询参数格式转换成其它类型的接口。
//
// 比如一个查询参数格式如下：
//  /path?state=locked
// 而实际上后端的将 state 表示为一个数值：
//  type State int8
//  const StateLocked State = 1
// 那么只要 State 实现 UnmarshalQueryer 接口，就可以实现将 locked 转换成 1 的能力。
//  func (s *State) UnmarshalQuery(data string) error {
//      if data == "locked" {
//          *s = StateLocked
//      }
//  }
type UnmarshalQueryer interface {
	// data 表示由查询参数传递过来的单个值。
	UnmarshalQuery(data string) error
}

// QueryObject 将查询参数解析到一个对象中。
//
// 返回的是每一个字段对应的错误信息。
//
//
// struct tag
//
// 通过 struct tag 的方式将查询参数与结构体中的字段进行关联。
// struct tag 的格式如下：
//  `query:"name,default"`
// 其中 name 为对应查询参数的名称，或是为空则采用字段本身的名称；
// default 表示在没有参数的情况下，采用的默认值，可以为空。
//
//
// 数组：
//
// 如果字段表示的是切片，那么查询参数的值，将以半角逗号作为分隔符进行转换写入到切片中。
// struct tag 中的默认值，也可以指定多个：
//  type Object struct {
//      Slice []int `query:"name,1,2"`
//  }
// 以上内容，在没有指定参数的情况下，Slice 会被指定为 []int{1,2}
//
//
// 默认值：
//
// 默认值可以通过 struct tag 指定，也可以通过在初始化对象时，另外指定：
//  obj := &Object{
//      Slice: []int{3,4,5}
//  }
// 以上内容，在不传递参数时，会采用 []int{3,4,5} 作为其默认值，而不是 struct tag
// 中指定的 []int{1,2}。
func (ctx *Context) QueryObject(v interface{}) (errors map[string]string) {
	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	ret := make(map[string]string, rval.NumField())
	parseField(ctx.Request, rval, ret)

	// 接口在转换完成之后调用。
	if q, ok := v.(QuerySanitizer); ok {
		q.QuerySanitize(ret)
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

		vf := rval.Field(i)

		switch tf.Type.Kind() {
		case reflect.Slice:
			parseFieldSlice(r, errors, tf, vf)
		case reflect.Array:
			parseFieldArray(r, errors, tf, vf)
		case reflect.Ptr, reflect.Chan, reflect.Func:
			continue LOOP
		default:
			parseFieldValue(r, errors, tf, vf)
		}
	} // end for
}

func parseFieldValue(r *http.Request, errors map[string]string, tf reflect.StructField, vf reflect.Value) {
	name, def := getQueryTag(tf)
	if name == "" {
		return
	}

	val := r.FormValue(name)
	if val == "" {
		if vf.Interface() != reflect.Zero(tf.Type).Interface() {
			return
		}
		val = def
	}

	if val == "" { // 依然是空值
		return
	}

	if q, ok := vf.Addr().Interface().(UnmarshalQueryer); ok {
		if err := q.UnmarshalQuery(val); err != nil {
			errors[name] = err.Error()
			return
		}
	} else if err := conv.Value(val, vf); err != nil {
		errors[name] = err.Error()
		return
	}
}

func parseFieldSlice(r *http.Request, errors map[string]string, tf reflect.StructField, vf reflect.Value) {
	name, def := getQueryTag(tf)
	if name == "" {
		return
	}

	val := r.FormValue(name)

	if val == "" {
		if vf.Len() > 0 { // 有默认值，则采用默认值
			return
		}
		val = def
	}

	if val == "" { // 依然是空值
		return
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
		if q, ok := elem.Interface().(UnmarshalQueryer); ok {
			if err := q.UnmarshalQuery(v); err != nil {
				errors[name] = err.Error()
				return
			}
		} else if err := conv.Value(v, elem); err != nil {
			errors[name] = err.Error()
			return
		}
		vf.Set(reflect.Append(vf, elem.Elem()))
	}
}

func parseFieldArray(r *http.Request, errors map[string]string, tf reflect.StructField, vf reflect.Value) {
	name, def := getQueryTag(tf)
	if name == "" {
		return
	}

	val := r.FormValue(name)

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
		if q, ok := elem.Interface().(UnmarshalQueryer); ok {
			if err := q.UnmarshalQuery(v); err != nil {
				errors[name] = err.Error()
				return
			}
		}
		if err := conv.Value(v, elem); err != nil {
			errors[name] = err.Error()
			return
		}
	}
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
