// SPDX-License-Identifier: MIT

package text

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/content"
	"github.com/issue9/web/content/text/testobject"
)

var (
	_ content.MarshalFunc   = Marshal
	_ content.UnmarshalFunc = Unmarshal
)

func TestTextMarshal(t *testing.T) {
	a := assert.New(t)

	v := "123"
	data, err := Marshal(v)
	a.NotError(err).Equal(string(data), v)

	data, err = Marshal([]rune(v))
	a.NotError(err).Equal(string(data), v)

	data, err = Marshal([]byte(v))
	a.NotError(err).Equal(string(data), v)

	// 实现 TextMarshaler 的对象
	data, err = Marshal(&testobject.TextObject{Name: "test", Age: 5})
	a.NotError(err).NotNil(data).Equal(string(data), "test,5")

	// 未实现 TextMarshaler 接口的对象
	data, err = Marshal(&struct{}{})
	a.Error(err).Nil(data)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	v1 := &testobject.TextObject{}
	a.NotError(Unmarshal([]byte("test,5"), v1))
	a.Equal(v1.Name, "test").Equal(v1.Age, 5)

	v2 := &struct{}{}
	a.Error(Unmarshal(nil, v2))

	v3 := ""
	a.NotError(Unmarshal([]byte("v3"), &v3))
	a.Equal(v3, "v3")

	v4 := []byte{}
	a.NotError(Unmarshal([]byte("v4"), &v4))
	a.Equal(v4, []byte("v4"))

	v5 := []rune{}
	a.NotError(Unmarshal([]byte("v5"), &v5))
	a.Equal(v5, []rune("v5"))
}
