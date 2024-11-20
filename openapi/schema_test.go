// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
)

func TestParameter_valid(t *testing.T) {
	a := assert.New(t, false)

	p := &Parameter{}
	err := p.valid(false)
	a.Equal(err.Field, "Name")

	p = &Parameter{Name: "p1"}
	err = p.valid(false)
	a.Equal(err.Field, "Schema")

	p = &Parameter{Name: "p1", Schema: &Schema{}}
	err = p.valid(false)
	a.Equal(err.Field, "Schema.Type")

	p = &Parameter{Name: "p1", Schema: &Schema{Type: TypeString}}
	a.NotError(p.valid(false))

	p = &Parameter{Name: "p1", Schema: &Schema{Type: TypeObject}}
	err = p.valid(false)
	a.Equal(err.Field, "Schema")
}

func TestParameter_addComponents(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	p := &Parameter{}
	p.addComponents(d.components, InPath)
	a.Empty(d.paths)

	p = &Parameter{Schema: &Schema{Type: TypeString}}
	p.addComponents(d.components, InPath)
	a.Empty(d.paths)

	p = &Parameter{Schema: &Schema{Type: TypeString}, Ref: &Ref{Ref: "string"}}
	p.addComponents(d.components, InPath)
	a.Equal(d.components.paths["string"], p)
}

func TestSchema_addComponents(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	s := &Schema{}
	s.addComponents(d.components)
	a.Empty(d.components.schemas)

	s = &Schema{Type: TypeString}
	s.addComponents(d.components)
	a.Empty(d.components.schemas)

	s1 := &Schema{Type: TypeString, Ref: &Ref{Ref: "t1"}}
	s1.addComponents(d.components)
	a.Length(d.components.schemas, 1).Equal(d.components.schemas["t1"], s1)

	// 同名不会再添加
	s2 := &Schema{Type: TypeString, Ref: &Ref{Ref: "t1"}}
	s2.addComponents(d.components)
	a.Length(d.components.schemas, 1).Equal(d.components.schemas["t1"], s1)

	s2 = &Schema{Type: TypeString, Ref: &Ref{Ref: "t2"}, Items: s1}
	s2.addComponents(d.components)
	a.Length(d.components.schemas, 2).Equal(d.components.schemas["t2"], s2)

	s3 := &Schema{Type: TypeString, Ref: &Ref{Ref: "t3"}, Items: &Schema{Type: TypeNumber, Ref: &Ref{Ref: "t4"}}}
	s3.addComponents(d.components)
	a.Length(d.components.schemas, 4).
		Equal(d.components.schemas["t3"], s3).
		Equal(d.components.schemas["t4"].Type, TypeNumber)
}

func TestSchema_valid(t *testing.T) {
	a := assert.New(t, false)

	s := &Schema{}
	err := s.valid(false)
	a.Equal(err.Field, "Type")

	s = &Schema{Type: TypeString}
	a.NotError(s.valid(false))

	s = &Schema{Type: TypeString, Properties: map[string]*Schema{"f1": {Type: TypeString}}}
	err = s.valid(false)
	a.Equal(err.Field, "Type")
	s = &Schema{Type: TypeObject, Properties: map[string]*Schema{"f1": {}}}
	err = s.valid(false)
	a.Equal(err.Field, "Properties[f1].Type")
	s = &Schema{Type: TypeObject, Properties: map[string]*Schema{}} // 空对象是合法的
	a.Nil(s.valid(false))

	s = &Schema{Type: TypeString, Items: &Schema{Type: TypeString}}
	err = s.valid(false)
	a.Equal(err.Field, "Type")
	s = &Schema{Type: TypeArray}
	err = s.valid(false)
	a.Equal(err.Field, "Type")

	s = &Schema{Type: TypeString, AllOf: []*Schema{{}}}
	err = s.valid(false)
	a.Equal(err.Field, "AllOf[0].Type")

	s = &Schema{Type: TypeString, OneOf: []*Schema{{}}}
	err = s.valid(false)
	a.Equal(err.Field, "OneOf[0].Type")

	s = &Schema{Type: TypeString, AnyOf: []*Schema{{Type: TypeString}, {}}}
	err = s.valid(false)
	a.Equal(err.Field, "AnyOf[1].Type")
}

func TestSchema_build(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	var s *Schema
	a.Nil(s.build(p))

	s = &Schema{Ref: &Ref{Ref: "ref"}, Type: TypeArray}
	sr := s.build(p)
	a.Equal(sr.ref.Ref, "#/components/schemas/ref").Nil(sr.obj)

	s = &Schema{Type: TypeArray, Description: web.Phrase("lang")}
	sr = s.build(p)
	a.Nil(sr.ref).NotNil(sr.obj).
		Equal(sr.obj.Type, TypeArray).
		Equal(sr.obj.Description, "简体")
}

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
