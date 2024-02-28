// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package gob

import (
	"bytes"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

var (
	_ web.MarshalFunc   = Marshal
	_ web.UnmarshalFunc = Unmarshal
)

func TestGOB(t *testing.T) {
	a := assert.New(t, false)

	str1 := "123"
	data, err := Marshal(nil, str1)
	a.NotError(err)
	var str2 string
	a.NotError(Unmarshal(bytes.NewBuffer(data), &str2))
	a.Equal(str2, str1)

	type gObject struct {
		V  int
		PV *int
	}

	v := 5
	obj1 := &gObject{V: 22, PV: &v}
	data, err = Marshal(nil, obj1)
	a.NotError(err)
	obj2 := &gObject{}
	a.NotError(Unmarshal(bytes.NewBuffer(data), obj2))
	a.Equal(obj2, obj1)

	data, err = Marshal(nil, nil)
	a.Error(err).Nil(data)
}
