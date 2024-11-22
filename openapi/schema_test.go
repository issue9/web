// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

func TestDocument_NewSchema(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	s := d.newSchema(reflect.TypeFor[int]())
	a.Equal(s.Type, TypeInteger).
		Nil(s.Ref)

	s = d.newSchema(reflect.TypeFor[map[string]float32]())
	a.Equal(s.Type, TypeObject).
		Nil(s.Ref).
		NotNil(s.AdditionalProperties).
		Equal(s.AdditionalProperties.Type, TypeNumber)

	s = d.newSchema(reflect.ValueOf(&object{}).Type())
	a.Equal(s.Type, TypeObject).
		NotZero(s.Ref.Ref).
		Length(s.Properties, 3).
		Equal(s.Properties["id"].Type, TypeInteger).
		Equal(s.Properties["id"].XML.Name, "Id").
		Nil(s.Properties["Items"].XML).
		Equal(s.Properties["Items"].Type, TypeArray).
		NotZero(s.Properties["Items"].Items.Ref.Ref) // 引用了 object
}
