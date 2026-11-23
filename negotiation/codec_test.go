// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation

import (
	"bytes"
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/negotiation/codectest"
)

func TestNewCBOR(t *testing.T) {
	a := assert.New(t, false)
	codectest.Test(a, NewCBOR())
}

func TestNewGOB(t *testing.T) {
	a := assert.New(t, false)
	mt := NewGOB()
	a.NotNil(mt)

	str1 := "123"
	data, err := mt.Marshal(nil, str1)
	a.NotError(err)
	var str2 string
	a.NotError(mt.Unmarshal(bytes.NewBuffer(data), &str2))
	a.Equal(str2, str1)

	type gObject struct {
		V  int
		PV *int
	}

	v := 5
	obj1 := &gObject{V: 22, PV: &v}
	data, err = mt.Marshal(nil, obj1)
	a.NotError(err)
	obj2 := &gObject{}
	a.NotError(mt.Unmarshal(bytes.NewBuffer(data), obj2))
	a.Equal(obj2, obj1)

	data, err = mt.Marshal(nil, nil)
	a.Error(err).Nil(data)

	codectest.Test(a, mt)
}

func TestNewJSON(t *testing.T) {
	a := assert.New(t, false)
	codectest.Test(a, NewJSON())
}

func TestNewXML(t *testing.T) {
	a := assert.New(t, false)
	codectest.Test(a, NewXML())
}

func TestNewYAML(t *testing.T) {
	a := assert.New(t, false)
	codectest.Test(a, NewYAML())
}
