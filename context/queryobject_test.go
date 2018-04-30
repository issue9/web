// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/issue9/assert"
)

type testQueryString struct {
	String  string   `query:"string,str1,str2"`
	Strings []string `query:"strings,str1,str2"`
}

type testQueryObject struct {
	testQueryString
	Int    int       `query:"int,1"`
	Floats []float64 `query:"floats,1.1,2.2"`
	Array  [2]int    `query:"array,1,2"`
}

func TestParseField(t *testing.T) {
	a := assert.New(t)

	errors := map[string]string{}
	r := httptest.NewRequest(http.MethodGet, "/q?string=str&strings=s1,s2", nil)
	data := &testQueryObject{}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.String, "str").
		Equal(data.Strings, []string{"s1", "s2"}).
		Equal(data.Int, 1). // 默认值
		Equal(data.Floats, []float64{1.1, 2.2}).
		Equal(data.Array, []int{1, 2})

	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=1,1.1&array=10,20,3&int=5&strings=s1", nil)
	data = &testQueryObject{}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Empty(errors)
	a.Equal(data.String, "str1,str2").
		Equal(data.Floats, []float64{1.0, 1.1}).
		Equal(data.Strings, []string{"s1"}).
		Equal(data.Array, []int{10, 20}) // array 多传的值，不会进来。

	// 出错时的处理
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=str,1.1&array=10&int=5&strings=s1", nil)
	data = &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.True(errors["floats"] != "") // floats 解析会出错
	a.Equal(data.String, "str1,str2").
		Empty(data.Floats).
		Equal(data.Strings, []string{"s1"}).
		Equal(data.Array, []int{10, 0}) // array 少传的值，第二个值采用默认值。
}

func TestParseField_slice(t *testing.T) {
	a := assert.New(t)

	// 指定了默认值，也指定了参数。则以参数优先
	errors := map[string]string{}
	r := httptest.NewRequest(http.MethodGet, "/q?floats=11.1", nil)
	data := &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Floats, []float64{11.1})

	// 指定了默认值，指定空参数，则使用默认值
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?floats=", nil)
	data = &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Floats, []float64{3.3, 4.4})

	// 指定了默认值，未指定数，则使用默认值
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q", nil)
	data = &testQueryObject{
		Floats: []float64{3.3, 4.4},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Floats, []float64{3.3, 4.4})

	// 都未指定，则使用 struct tag 中的默认值
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q", nil)
	data = &testQueryObject{}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Floats, []float64{1.1, 2.2})
}

func TestParseField_array(t *testing.T) {
	a := assert.New(t)

	// 指定了默认值，也指定了参数。则以参数优先
	errors := map[string]string{}
	r := httptest.NewRequest(http.MethodGet, "/q?array=11,1", nil)
	data := &testQueryObject{
		Array: [2]int{2, 2},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Array, []int{11, 1})

	// 指定了默认值，指定空参数，则使用默认值
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q?array=", nil)
	data = &testQueryObject{
		Array: [2]int{2, 2},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Array, []int{2, 2})

	// 指定了默认值，未指定数，则使用默认值
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q", nil)
	data = &testQueryObject{
		Array: [2]int{2, 2},
	}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Array, []int{2, 2})

	// 都未指定，则使用 struct tag 中的默认值
	errors = map[string]string{}
	r = httptest.NewRequest(http.MethodGet, "/q", nil)
	data = &testQueryObject{}
	parseField(r, reflect.ValueOf(data).Elem(), errors)
	a.Equal(data.Array, []int{1, 2})
}

func TestGetQueryTag(t *testing.T) {
	a := assert.New(t)

	test := func(tag, name, def string) {
		field := reflect.StructField{
			Tag: reflect.StructTag(tag),
		}
		n, d := getQueryTag(field)
		a.Equal(n, name).
			Equal(d, def)
	}

	test(`query:"name,def"`, "name", "def")
	test(`query:",def"`, "", "def")
	test(`query:"name,"`, "name", "")
	test(`query:"name"`, "name", "")
	test(`query:"name,1,2"`, "name", "1,2")
	test(`query:"name,1,2,"`, "name", "1,2,")
	test(`query:"-"`, "", "")
}
