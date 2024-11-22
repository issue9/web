// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

type schemaObject1 struct {
	object
	Root string
	T    time.Time
	X    string  `openapi:"-"`
	Y    string  `openapi:"integer"`
	Z    *object `openapi:"string,date"`
}

type schemaObject2 struct {
	schemaObject1
}

func TestDocument_newSchema(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	s := d.newSchema(reflect.TypeFor[int]())
	a.Equal(s.Type, TypeInteger).
		Nil(s.Ref)

	s = d.newSchema(reflect.TypeFor[[]int]())
	a.Equal(s.Type, TypeArray).
		Equal(s.Items.Type, TypeInteger)

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

	s = d.newSchema(reflect.ValueOf(schemaObject1{}).Type())
	a.Equal(s.Type, TypeObject).
		NotZero(s.Ref.Ref).
		Length(s.Properties, 7).
		Equal(s.Properties["id"].Type, TypeInteger).
		Equal(s.Properties["Root"].Type, TypeString).
		Equal(s.Properties["T"].Type, TypeString).
		Equal(s.Properties["T"].Format, FormatDateTime).
		Equal(s.Properties["Y"].Type, TypeInteger).
		Equal(s.Properties["Z"].Type, TypeString).
		Equal(s.Properties["Z"].Format, FormatDate)

	s = d.newSchema(reflect.ValueOf(schemaObject2{}).Type())
	a.Equal(s.Type, TypeObject).
		NotZero(s.Ref.Ref).
		Length(s.Properties, 7).
		Equal(s.Properties["id"].Type, TypeInteger).
		Equal(s.Properties["Root"].Type, TypeString).
		Equal(s.Properties["T"].Type, TypeString).
		Equal(s.Properties["T"].Format, FormatDateTime)
}

func TestSchema_isBasicType(t *testing.T) {
	a := assert.New(t, false)

	s := NewSchema(reflect.TypeFor[int](), nil, nil)
	a.True(s.isBasicType())

	s = NewSchema(reflect.TypeFor[object](), nil, nil)
	a.False(s.isBasicType())

	s = NewSchema(reflect.TypeFor[[]string](), nil, nil)
	a.True(s.isBasicType())
}
