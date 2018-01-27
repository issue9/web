// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding/test"
)

var (
	_ MarshalFunc   = TextMarshal
	_ UnmarshalFunc = TextUnmarshal
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
	data, err = TextMarshal(&test.TextObject{Name: "test", Age: 5})
	a.NotError(err).NotNil(data).Equal(string(data), "test,5")

	// 未实现 TextMarshaler 接口的对象
	data, err = TextMarshal(&struct{}{})
	a.Equal(err, ErrUnsupportedMarshal).Nil(data)
}

func TestTextUnmarshal(t *testing.T) {
	a := assert.New(t)

	v1 := &test.TextObject{}
	a.NotError(TextUnmarshal([]byte("test,5"), v1))
	a.Equal(v1.Name, "test").Equal(v1.Age, 5)

	v2 := &struct{}{}
	a.Equal(TextUnmarshal(nil, v2), ErrUnsupportedMarshal)
}
