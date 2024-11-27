// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
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

func TestOfSchema(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		AllOfSchema(nil, nil)
	}, "参数 v 必不可少")

	s := AnyOfSchema(web.Phrase("lang"), nil, "1", 0)
	a.Length(s.AnyOf, 2).
		Empty(s.Type).
		Equal(s.AnyOf[0].Type, TypeString).
		Equal(s.AnyOf[0].Default, "1").
		Equal(s.AnyOf[1].Type, TypeInteger).
		Nil(s.AnyOf[1].Default)

	s = OneOfSchema(web.Phrase("lang"), nil, true, uint(2))
	a.Length(s.OneOf, 2).
		Empty(s.Type).
		Equal(s.OneOf[0].Type, TypeBoolean).
		Equal(s.OneOf[0].Default, true).
		Equal(s.OneOf[1].Type, TypeInteger)

	s = AllOfSchema(web.Phrase("lang"), nil, "1", 2)
	a.Length(s.AllOf, 2).
		Empty(s.Type).
		Equal(s.AllOf[0].Type, TypeString).
		Equal(s.AllOf[1].Type, TypeInteger)
}

func TestDocument_newSchema(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	s := d.newSchema(5)
	a.Equal(s.Type, TypeInteger).
		Nil(s.Ref).
		Equal(s.Default, 5)

	s = d.newSchema([]int{5, 6})
	a.Equal(s.Type, TypeArray).
		Equal(s.Items.Type, TypeInteger).
		Equal(s.Default, []int{5, 6})

	s = d.newSchema(map[string]float32{"1": 3.2})
	a.Equal(s.Type, TypeObject).
		Nil(s.Ref).
		NotNil(s.AdditionalProperties).
		Equal(s.AdditionalProperties.Type, TypeNumber)

	s = d.newSchema(&object{})
	a.Equal(s.Type, TypeObject).
		NotZero(s.Ref.Ref).
		Length(s.Properties, 3).
		Equal(s.Properties["id"].Type, TypeInteger).
		Equal(s.Properties["id"].XML.Name, "Id").
		Nil(s.Properties["Items"].XML).
		Equal(s.Properties["Items"].Type, TypeArray).
		NotZero(s.Properties["Items"].Items.Ref.Ref) // 引用了 object

	s = d.newSchema(schemaObject1{})
	a.Equal(s.Type, TypeObject).
		NotZero(s.Ref.Ref).
		Nil(s.Default).
		Length(s.Properties, 7).
		Equal(s.Properties["id"].Type, TypeInteger).
		Equal(s.Properties["Root"].Type, TypeString).
		Equal(s.Properties["T"].Type, TypeString).
		Equal(s.Properties["T"].Format, FormatDateTime).
		Equal(s.Properties["Y"].Type, TypeInteger).
		Equal(s.Properties["Z"].Type, TypeString).
		Equal(s.Properties["Z"].Format, FormatDate)

	s = d.newSchema(schemaObject2{})
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

	s := NewSchema(5, nil, nil)
	a.True(s.isBasicType())

	s = NewSchema(object{}, nil, nil)
	a.False(s.isBasicType())

	s = NewSchema([]string{}, nil, nil)
	a.True(s.isBasicType())
}
