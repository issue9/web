// SPDX-License-Identifier: MIT

// Package form 用于处理 www-form-urlencoded 编码
//
//  func read(ctx *web.Context) {
//      vals := urls.Values{}
//      ctx.Read(vals)
//  }
//
//  func write(ctx *web.Context) {
//      vals := urls.Values{}
//      vals.Add("name", "caixw")
//      ctx.Render(http.StatusOK, vals, nil)
//  }
//
//
// form
//
// 用户可以通过定义 form 标签自定义输出的名称，比如：
//  type Username struct {
//      Name string `form:"name"`
//      Age int
//  }
// 转换成 form-data 可能是以下样式：
//  name=jjj&age=18
// 该方式对数据类型有一定限制：
//  1. 如果是 map 类型，要求键名类型必须为 string；
//  2. 如果是 array 或是 slice，则要求元素类型必须是 go 的基本数据类型，不能是 struct 类型；
//
//
// 接口
//
// 对于复杂类型，用户可以自定义实现 Marshaler 和 Unmarshaler 接口进行编解码，
// 其功能与用户与 encoding/json 中的 Marshaler 和 Unmarshaler 接口相似。
package form

import (
	"encoding"
	"net/url"
)

// Mimetype 当前编码的媒体类型
const Mimetype = "application/x-www-form-urlencoded"

// Marshaler 将一个普通对象转换成 form 类型
type Marshaler interface {
	MarshalForm() ([]byte, error)
}

// Unmarshaler 将 form 类型转换成一个对象
type Unmarshaler interface {
	UnmarshalForm([]byte) error
}

// Marshal 针对 www-form-urlencoded 内容的 content.MarshalFunc 实现
//
// 按以下顺序解析内容：
//  - 如果实现 Marshaler 接口，则调用该接口；
//  - 如果实现 encoding.TextMarshaler 接口，则调用该接口；
//  - 如果是 url.Values 对象，则调用 url.Values.Encode() 解析；
//  - 否则将对象的字段与 form-data 中的数据进行对比，可以使用 form 指定字段名。
func Marshal(v interface{}) ([]byte, error) {
	if m, ok := v.(Marshaler); ok {
		return m.MarshalForm()
	}

	if m, ok := v.(encoding.TextMarshaler); ok {
		return m.MarshalText()
	}

	if vals, ok := v.(url.Values); ok {
		return []byte(vals.Encode()), nil
	}

	vals, err := marshal(v)
	if err != nil {
		return nil, err
	}

	return []byte(vals.Encode()), nil
}

// Unmarshal 针对 www-form-urlencoded 内容的 content.UnmarshalFunc 实现
//
// 按以下顺序解析内容：
//  - 如果实现 Unmarshaler 接口，则调用该接口；
//  - 如果是 url.Values 对象，则依次赋值每个对象；
//  - 否则将对象的字段与 form-data 中的数据进行对比，可以使用 form 指定字段名。
func Unmarshal(data []byte, v interface{}) error {
	if m, ok := v.(Unmarshaler); ok {
		return m.UnmarshalForm(data)
	}

	vals, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	if obj, ok := v.(url.Values); ok {
		for k, v := range vals {
			obj[k] = v
		}
		return nil
	}

	return unmarshal(vals, v)
}
