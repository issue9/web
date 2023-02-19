// SPDX-License-Identifier: MIT

package html

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/server"
)

var (
	_ server.MarshalFunc   = Marshal
	_ server.UnmarshalFunc = Unmarshal

	_ Marshaler = &problems.RFC7807[*server.Context]{}
)

func TestGetName(t *testing.T) {
	a := assert.New(t, false)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}
	type obj2 struct {
		HTMLName struct{}
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

	name, v = getName(server.RFC7807Builder("id", 500, "title", "detail"))
	a.Equal(name, "problem").
		Equal(v, map[string]any{
			"type":   "id",
			"title":  "title",
			"status": 500,
		})

	a.PanicString(func() {
		getName(map[string]string{})
	}, "text/html 不支持输出当前类型 map")
}
