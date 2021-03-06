// SPDX-License-Identifier: MIT

package mimetypetest

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/content"
)

var (
	_ content.MarshalFunc   = TextMarshal
	_ content.UnmarshalFunc = TextUnmarshal
)

func TestTextMarshal(t *testing.T) {
	a := assert.New(t)

	v := "123"
	data, err := TextMarshal(v)
	a.NotError(err).Equal(string(data), v)

	data, err = TextMarshal([]rune(v))
	a.NotError(err).Equal(string(data), v)

	data, err = TextMarshal([]byte(v))
	a.NotError(err).Equal(string(data), v)

	// 实现 TextMarshaler 的对象
	data, err = TextMarshal(&TextObject{Name: "test", Age: 5})
	a.NotError(err).NotNil(data).Equal(string(data), "test,5")

	// 未实现 TextMarshaler 接口的对象
	data, err = TextMarshal(&struct{}{})
	a.Error(err).Nil(data)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	v1 := &TextObject{}
	a.NotError(TextUnmarshal([]byte("test,5"), v1))
	a.Equal(v1.Name, "test").Equal(v1.Age, 5)

	v2 := &struct{}{}
	a.Error(TextUnmarshal(nil, v2))
}
