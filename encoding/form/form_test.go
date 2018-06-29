// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package form

import (
	"net/url"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
)

var (
	_ encoding.MarshalFunc   = Marshal
	_ encoding.UnmarshalFunc = Unmarshal
)

var formString = "friend=Jess&friend=Sarah&friend=Zoe&name=Ava"

func init() {
}

func TestMarshal(t *testing.T) {
	a := assert.New(t)

	formObject := url.Values{}
	data, err := Marshal(formObject)
	a.NotError(err)
	a.NotNil(data). // 非 nil
			Empty(data) // 但长度为 0

	formObject.Set("name", "Ava")
	formObject.Add("friend", "Jess")
	formObject.Add("friend", "Sarah")
	formObject.Add("friend", "Zoe")
	data, err = Marshal(formObject)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), formString)

	// 非 url.Values 类型
	data, err = Marshal(&struct{}{})
	a.ErrorType(err, errInvalidType).Nil(data)
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	v := url.Values{}
	a.NotError(Unmarshal(nil, v))
	a.Equal(len(v), 0)

	v = url.Values{}
	a.NotError(Unmarshal([]byte{}, v))
	a.Equal(len(v), 0)

	v = url.Values{}
	a.Error(Unmarshal([]byte("%"), v))

	a.ErrorType(Unmarshal([]byte(formString), &struct{}{}), errInvalidType)

	v = url.Values{}
	a.NotError(Unmarshal([]byte(formString), v))
	a.Equal(len(v), 2)
	a.Equal(v.Get("name"), "Ava")
	a.Equal(v.Get("friend"), "Jess")
	a.Equal(v["friend"], []string{"Jess", "Sarah", "Zoe"})
}
