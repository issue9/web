// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package codec

import (
	"bytes"
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/codec/mediatypetest"
)

func TestNewMeidaType(t *testing.T) {
	a := assert.New(t, false)

	mt1 := NewCBOR()
	mt2 := NewMediaType(mt1, "mt2", "")
	a.Equal(mt2.Name(false), "mt2").
		Equal(mt2.Name(true), "mt2").
		NotEqual(mt1.Name(false), mt2.Name(false))
}

func TestNewCBOR(t *testing.T) {
	a := assert.New(t, false)
	mediatypetest.Test(a, NewCBOR())
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

	mediatypetest.Test(a, mt)
}

func TestNewJSON(t *testing.T) {
	a := assert.New(t, false)
	mediatypetest.Test(a, NewJSON())
}

func TestNewXML(t *testing.T) {
	a := assert.New(t, false)
	mediatypetest.Test(a, NewXML())
}

func TestNewYAML(t *testing.T) {
	a := assert.New(t, false)
	mediatypetest.Test(a, NewYAML())
}
