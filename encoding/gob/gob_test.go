// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package gob

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
)

var (
	_ encoding.MarshalFunc   = Marshal
	_ encoding.UnmarshalFunc = Unmarshal
)

func TestGOB(t *testing.T) {
	a := assert.New(t)

	str1 := "123"
	data, err := Marshal(str1)
	a.NotError(err)
	var str2 string
	a.NotError(Unmarshal(data, &str2))
	a.Equal(str2, str1)

	type gobject struct {
		V int
	}

	obj1 := &gobject{V: 22}
	data, err = Marshal(obj1)
	a.NotError(err)
	obj2 := &gobject{}
	a.NotError(Unmarshal(data, obj2))
	a.Equal(obj2, obj1)
}
