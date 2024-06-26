// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package html

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

var (
	_ web.MarshalFunc   = Marshal
	_ web.UnmarshalFunc = Unmarshal
	_ Marshaler         = &web.Problem{}
)

func TestGetName(t *testing.T) {
	a := assert.New(t, false)

	type obj struct {
		XMLName struct{} `html:"t"`
	}
	type obj2 struct {
		XMLName struct{}
	}

	type obj3 struct{}

	type obj4 map[string]string

	name, v := getName(&obj{})
	a.Equal(name, "t").Empty(v) // 指针类型的 v

	name, v = getName(&obj2{})
	a.Equal(name, "obj2").Zero(v)

	name, v = getName(&obj3{})
	a.Equal(name, "obj3").Zero(v)

	name, v = getName(&obj4{})
	a.Equal(name, "obj4").Empty(v)

	a.PanicString(func() {
		getName(map[string]string{})
	}, "text/html 不支持输出当前类型 map")
}
