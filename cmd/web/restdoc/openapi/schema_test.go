// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v4"
)

func TestNewSchemaRef(t *testing.T) {
	a := assert.New(t, false)
	s := openapi3.NewSchema()

	ref := NewSchemaRef("", s)
	a.Equal(ref.Ref, "")

	ref = NewSchemaRef("ref_id", s)
	a.Equal(ref.Ref, ComponentSchemaPrefix+"ref_id")

	ref = NewSchemaRef(ComponentSchemaPrefix+"ref_id", s)
	a.Equal(ref.Ref, ComponentSchemaPrefix+"ref_id")
}

func TestNewArraySchemaRef(t *testing.T) {
	a := assert.New(t, false)
	s := openapi3.NewSchema()

	ref := NewSchemaRef("", s)
	arr := NewArraySchemaRef(ref)
	a.Equal(arr.Ref, "")

	ref = NewSchemaRef("id", s)
	arr = NewArraySchemaRef(ref)
	a.Equal(arr.Ref, "")

	ref = NewSchemaRef(ComponentSchemaPrefix+"id", s)
	arr = NewArraySchemaRef(ref)
	a.Equal(arr.Ref, "")
}

func TestOpenapi_Schema(t *testing.T) {
	a := assert.New(t, false)
	o := New("3.0.0")

	ref := NewSchemaRef("", nil)
	o.AddSchema(ref)
	a.Length(o.doc.Components.Schemas, 0) // 空的 refID，添加不成功

	ref = NewSchemaRef("abc", nil)
	o.AddSchema(ref)
	a.Length(o.doc.Components.Schemas, 1).
		Equal(ref.Ref, ComponentSchemaPrefix+"abc")

	v1, found := o.GetSchema("abc")
	a.True(found).NotNil(v1)
	v2, found := o.GetSchema(ComponentSchemaPrefix + "abc")
	a.True(found).NotNil(v2)
	a.Equal(v1, v2).
		Equal(v1, v2). // v1,v2 指向同一个对象
		Equal(v1.Ref, ComponentSchemaPrefix+"abc")

	// 同名，但都是 nil
	a.NotPanic(func() {
		ref = NewSchemaRef(ComponentSchemaPrefix+"abc", nil)
		o.AddSchema(ref)
	})

	a.PanicString(func() {
		ref = NewSchemaRef(ComponentSchemaPrefix+"abc", &openapi3.Schema{})
		o.AddSchema(ref)
	}, "添加同名的对象 abc")
}

func TestNewAttrSchemaRef(t *testing.T) {
	a := assert.New(t, false)

	ref := openapi3.NewSchemaRef("ref", openapi3.NewBoolSchema())
	ref2 := NewAttrSchemaRef(ref, "", "", nil, false)
	a.Equal(ref2, ref)

	ref2 = NewAttrSchemaRef(ref, "", "123", nil, false)
	a.NotEqual(ref2, ref).
		Equal(ref2.Value.AllOf[0].Value, ref.Value).
		Equal(ref2.Value.Description, "123")

	ref2 = NewAttrSchemaRef(ref, "", "123", nil, true)
	a.NotEqual(ref2, ref).
		Equal(ref2.Value.AllOf[0].Value, ref.Value).
		Equal(ref2.Value.Description, "123").
		True(ref2.Value.Nullable)
}
