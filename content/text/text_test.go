// SPDX-License-Identifier: MIT

package text

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/content"
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
	data, err = Marshal(&TestObject{Name: "test", Age: 5})
	a.NotError(err).NotNil(data).Equal(string(data), "test,5")

	// 未实现 TextMarshaler 接口的对象
	data, err = Marshal(&struct{}{})
	a.Error(err).Nil(data)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	v1 := &TestObject{}
	a.NotError(Unmarshal([]byte("test,5"), v1))
	a.Equal(v1.Name, "test").Equal(v1.Age, 5)

	v2 := &struct{}{}
	a.Error(Unmarshal(nil, v2))
}
