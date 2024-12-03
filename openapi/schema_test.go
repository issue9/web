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

type State int8

func (ss State) OpenAPISchema(s *Schema) {
	s.Type = TypeString
	s.Enum = []any{"1", "2"}
}

type schemaObject1 struct {
	object
	Root string
	T    time.Time
	X    string  `openapi:"-"`
	Y    string  `openapi:"integer"`
	Z    *object `openapi:"string,date"`

	S1 State  `openapi:"integer" json:"s1"`
	S2 *State `comment:"s2" json:"s2" xml:"S2"`
	S3 State

	unExported bool
}

type schemaObject2 struct {
	schemaObject1
}

type schemaObject3 struct {
	X int
}

func (*schemaObject3) OpenAPISchema(s *Schema) {
	s.Type = TypeString
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

	a.Nil(d.newSchema(nil))

	s = d.newSchema(map[string]float32{"1": 3.2})
	a.Equal(s.Type, TypeObject).
		Nil(s.Ref).
		NotNil(s.AdditionalProperties).
		Equal(s.AdditionalProperties.Type, TypeNumber)

	// 指针实现 OpenAPISchema
	s = d.newSchema(schemaObject3{})
	a.Equal(s.Type, TypeObject)
	s = d.newSchema(&schemaObject3{})
	a.Equal(s.Type, TypeString)
	so := &schemaObject3{}
	s = d.newSchema(&so) // ** schemaObject{}
	a.Equal(s.Type, TypeObject)

	// 值类型实现 OpenAPISchema
	s = d.newSchema(State(5))
	a.Equal(s.Type, TypeString)
	sss := State(5)
	s = d.newSchema(&sss)
	a.Equal(s.Type, TypeString)
	ssss := &sss
	s = d.newSchema(&ssss) // ** State
	a.Equal(s.Type, TypeInteger)

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
		Length(s.Properties, 10).
		Equal(s.Properties["id"].Type, TypeInteger).
		Equal(s.Properties["Root"].Type, TypeString).
		Equal(s.Properties["T"].Type, TypeString).
		Equal(s.Properties["T"].Format, FormatDateTime).
		Equal(s.Properties["Y"].Type, TypeInteger).
		Equal(s.Properties["Z"].Type, TypeString).
		Equal(s.Properties["Z"].Format, FormatDate).
		Equal(s.Properties["s1"].Type, TypeInteger). // openapi 标签优先于 OpenAPISchema
		Equal(s.Properties["s2"].Type, TypeString).  // OpenAPISchema 接口
		Equal(s.Properties["s2"].XML.Name, "S2").
		Equal(s.Properties["s2"].Enum, []any{"1", "2"}).
		Equal(s.Properties["s2"].Description, web.Phrase("s2")). // 注释可正确获取
		Equal(s.Properties["S3"].Type, TypeString).
		Nil(s.Properties["S3"].XML).
		Equal(s.Properties["S3"].Enum, []any{"1", "2"})

	s = d.newSchema(schemaObject2{})
	a.Equal(s.Type, TypeObject).
		NotZero(s.Ref.Ref).
		Length(s.Properties, 10).
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
