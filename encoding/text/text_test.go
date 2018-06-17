// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package text_test

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
	"github.com/issue9/web/encoding/test"
	"github.com/issue9/web/encoding/text"
)

var (
	_ encoding.MarshalFunc   = text.Marshal
	_ encoding.UnmarshalFunc = text.Unmarshal
)

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	v := "123"
	data, err := text.Marshal(v)
	a.NotError(err).Equal(string(data), v)

	data, err = text.Marshal([]rune(v))
	a.NotError(err).Equal(string(data), v)

	data, err = text.Marshal([]byte(v))
	a.NotError(err).Equal(string(data), v)

	// 实现 TextMarshaler 的对象
	data, err = text.Marshal(&test.TextObject{Name: "test", Age: 5})
	a.NotError(err).NotNil(data).Equal(string(data), "test,5")

	// 未实现 TextMarshaler 接口的对象
	data, err = text.Marshal(&struct{}{})
	a.Error(err).Nil(data)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	v1 := &test.TextObject{}
	a.NotError(text.Unmarshal([]byte("test,5"), v1))
	a.Equal(v1.Name, "test").Equal(v1.Age, 5)

	v2 := &struct{}{}
	a.Error(text.Unmarshal(nil, v2))
}
