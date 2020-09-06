// SPDX-License-Identifier: MIT

package gob_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
)

var (
	_ mimetype.MarshalFunc   = gob.Marshal
	_ mimetype.UnmarshalFunc = gob.Unmarshal
)

func TestGOB(t *testing.T) {
	a := assert.New(t)

	str1 := "123"
	data, err := gob.Marshal(str1)
	a.NotError(err)
	var str2 string
	a.NotError(gob.Unmarshal(data, &str2))
	a.Equal(str2, str1)

	type gobject struct {
		V  int
		PV *int
	}

	v := 5
	obj1 := &gobject{V: 22, PV: &v}
	data, err = gob.Marshal(obj1)
	a.NotError(err)
	obj2 := &gobject{}
	a.NotError(gob.Unmarshal(data, obj2))
	a.Equal(obj2, obj1)

	data, err = gob.Marshal(nil)
	a.Error(err).Nil(data)
}
