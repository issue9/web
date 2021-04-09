// SPDX-License-Identifier: MIT

package gob

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/content"
)

var (
	_ content.MarshalFunc   = Marshal
	_ content.UnmarshalFunc = Unmarshal
)

func TestGOB(t *testing.T) {
	a := assert.New(t)

	str1 := "123"
	data, err := Marshal(str1)
	a.NotError(err)
	var str2 string
	a.NotError(Unmarshal(data, &str2))
	a.Equal(str2, str1)

	type gObject struct {
		V  int
		PV *int
	}

	v := 5
	obj1 := &gObject{V: 22, PV: &v}
	data, err = Marshal(obj1)
	a.NotError(err)
	obj2 := &gObject{}
	a.NotError(Unmarshal(data, obj2))
	a.Equal(obj2, obj1)

	data, err = Marshal(nil)
	a.Error(err).Nil(data)
}
